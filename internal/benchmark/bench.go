package benchmark

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/romeo-mz/lokai/internal/ollama"
)

// Result holds the outcome of a single model benchmark.
type Result struct {
	ModelTag         string        `json:"model_tag"`
	Prompt           string        `json:"prompt"`
	TokensGenerated  int           `json:"tokens_generated"`
	TotalDuration    time.Duration `json:"total_duration_ns"`
	LoadDuration     time.Duration `json:"load_duration_ns"`
	PromptEvalRate   float64       `json:"prompt_eval_tokens_per_sec"`
	EvalRate         float64       `json:"eval_tokens_per_sec"`
	TimeToFirstToken time.Duration `json:"time_to_first_token_ns"`
	Success          bool          `json:"success"`
	Error            string        `json:"error,omitempty"`
}

// FormattedSpeed returns a human-readable generation speed.
func (r *Result) FormattedSpeed() string {
	if !r.Success {
		return "failed"
	}
	return fmt.Sprintf("%.1f tok/s", r.EvalRate)
}

// FormattedTTFT returns a human-readable time-to-first-token.
func (r *Result) FormattedTTFT() string {
	if !r.Success {
		return "—"
	}
	return fmt.Sprintf("%.2fs", r.TimeToFirstToken.Seconds())
}

// FormattedTotal returns a human-readable total duration.
func (r *Result) FormattedTotal() string {
	if !r.Success {
		return "—"
	}
	return r.TotalDuration.Round(time.Millisecond).String()
}

// Options configures a benchmark run.
type Options struct {
	Prompt    string // Custom prompt (default: standard benchmark prompt)
	MaxTokens int    // Max tokens to generate (default: 128)
	Warmup    bool   // Run a warmup generation first (default: true)
}

var defaultPrompt = "Explain the concept of recursion in programming in exactly 100 words."

// Run benchmarks a single model and returns the result.
func Run(ctx context.Context, client *ollama.Client, modelTag string, opts Options) Result {
	if opts.Prompt == "" {
		opts.Prompt = defaultPrompt
	}
	if opts.MaxTokens <= 0 {
		opts.MaxTokens = 128
	}

	result := Result{
		ModelTag: modelTag,
		Prompt:   opts.Prompt,
	}

	// Check if model is installed.
	installed, err := client.IsModelInstalled(ctx, modelTag)
	if err != nil || !installed {
		result.Error = fmt.Sprintf("model %s is not installed", modelTag)
		return result
	}

	// Optional warmup.
	if opts.Warmup {
		_ = generate(ctx, client, modelTag, "Hi", 10)
	}

	// Actual benchmark.
	genResult := generate(ctx, client, modelTag, opts.Prompt, opts.MaxTokens)
	if genResult.err != nil {
		result.Error = genResult.err.Error()
		return result
	}

	result.Success = true
	result.TokensGenerated = genResult.tokensGenerated
	result.TotalDuration = genResult.totalDuration
	result.LoadDuration = genResult.loadDuration
	result.TimeToFirstToken = genResult.timeToFirstToken
	result.PromptEvalRate = genResult.promptEvalRate
	result.EvalRate = genResult.evalRate

	return result
}

// RunMultiple benchmarks multiple models sequentially.
func RunMultiple(ctx context.Context, client *ollama.Client, modelTags []string, opts Options, onResult func(Result)) []Result {
	var results []Result
	for _, tag := range modelTags {
		r := Run(ctx, client, tag, opts)
		results = append(results, r)
		if onResult != nil {
			onResult(r)
		}
	}
	return results
}

type genResult struct {
	tokensGenerated  int
	totalDuration    time.Duration
	loadDuration     time.Duration
	timeToFirstToken time.Duration
	promptEvalRate   float64
	evalRate         float64
	err              error
}

func generate(ctx context.Context, client *ollama.Client, model, prompt string, maxTokens int) genResult {
	start := time.Now()
	var firstTokenTime time.Time
	var sb strings.Builder
	tokenCount := 0
	var stats ollama.GenerateStats

	err := client.Generate(ctx, model, prompt, func(token string, done bool) {
		if tokenCount == 0 {
			firstTokenTime = time.Now()
		}
		sb.WriteString(token)
		tokenCount++
	}, &stats)

	totalDuration := time.Since(start)

	if err != nil {
		return genResult{err: err}
	}

	ttft := time.Duration(0)
	if !firstTokenTime.IsZero() {
		ttft = firstTokenTime.Sub(start)
	}

	// Use Ollama's native metrics when available; fall back to wall-clock estimates.
	evalRate := 0.0
	if stats.EvalDuration > 0 && stats.EvalCount > 0 {
		evalRate = float64(stats.EvalCount) / stats.EvalDuration.Seconds()
	} else {
		evalDuration := totalDuration - ttft
		if evalDuration.Seconds() > 0 && tokenCount > 0 {
			evalRate = float64(tokenCount) / evalDuration.Seconds()
		}
	}

	promptEvalRate := 0.0
	if stats.PromptEvalDuration > 0 && stats.PromptEvalCount > 0 {
		promptEvalRate = float64(stats.PromptEvalCount) / stats.PromptEvalDuration.Seconds()
	} else if ttft.Seconds() > 0 {
		// Rough estimate: prompt tokens ≈ words * 1.3
		promptTokens := float64(len(strings.Fields(prompt))) * 1.3
		promptEvalRate = promptTokens / ttft.Seconds()
	}

	return genResult{
		tokensGenerated:  tokenCount,
		totalDuration:    totalDuration,
		loadDuration:     stats.LoadDuration,
		timeToFirstToken: ttft,
		evalRate:         evalRate,
		promptEvalRate:   promptEvalRate,
	}
}
