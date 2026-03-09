// Package models — static model catalog.
//
// Hand-curated database of Ollama-compatible models with quality scores,
// VRAM estimates, and use-case mappings.
//
// Sources:
//
//	Ollama Library — https://ollama.com/search
//	Model cards    — https://huggingface.co
//	Benchmarks     — https://huggingface.co/spaces/open-llm-leaderboard/open_llm_leaderboard
package models

import "github.com/romeo-mz/lokai/internal/hardware"

// UseCase represents a user's intended AI task.
type UseCase = hardware.UseCase

const (
	UseCaseChat      = hardware.UseCaseChat
	UseCaseCode      = hardware.UseCaseCode
	UseCaseVision    = hardware.UseCaseVision
	UseCaseEmbedding = hardware.UseCaseEmbedding
	UseCaseReasoning = hardware.UseCaseReasoning
	UseCaseImage     = hardware.UseCaseImage
	UseCaseVideo     = hardware.UseCaseVideo
	UseCaseAudio     = hardware.UseCaseAudio
	UseCaseUnrestricted      = hardware.UseCaseUnrestricted
)

// Priority represents the user's performance preference.
type Priority string

const (
	PriorityQuality  Priority = "quality"  // Best quality (largest model that fits)
	PrioritySpeed    Priority = "speed"    // Fastest inference
	PriorityBalanced Priority = "balanced" // Balance of quality and speed
)

// ModelEntry represents a model in the catalog.
type ModelEntry struct {
	Name            string    `json:"name"`              // Human-readable name
	OllamaTag       string    `json:"ollama_tag"`        // e.g. "llama3.1:8b"
	Family          string    `json:"family"`            // e.g. "llama", "gemma", "qwen"
	ParameterSize   string    `json:"parameter_size"`    // e.g. "8B", "70B"
	ParameterCount  float64   `json:"parameter_count"`   // e.g. 8.0, 70.0 (in billions)
	QuantLevel      string    `json:"quant_level"`       // e.g. "Q4_K_M", "Q4_0"
	DiskSizeGB      float64   `json:"disk_size_gb"`      // Approximate download size
	EstimatedVRAMGB float64   `json:"estimated_vram_gb"` // VRAM needed at runtime
	UseCases        []UseCase `json:"use_cases"`         // Supported use cases
	Capabilities    []string  `json:"capabilities"`      // "tools", "vision", "thinking", etc.
	Quality         int       `json:"quality"`           // Quality score 1-100 (higher = better)
	Description     string    `json:"description"`       // Short description
	// IsExternal marks models that cannot be installed via "ollama pull"
	// (e.g. diffusion models that need ComfyUI, whisper.cpp, etc.).
	IsExternal bool `json:"is_external,omitempty"`
	// ExternalURL is the HuggingFace or project page for non-Ollama models.
	ExternalURL string `json:"external_url,omitempty"`
	// Pipeline is the required inference software (e.g. "comfyui", "whisper.cpp").
	Pipeline string `json:"pipeline,omitempty"`
}

// IsPullable returns true when the model can be installed via "ollama pull".
func (m ModelEntry) IsPullable() bool { return !m.IsExternal }

// Catalog is the built-in model database.
// Ordered roughly by quality within each category.
var Catalog = []ModelEntry{
	// ═══════════════════════════════════════════
	// GENERAL CHAT — from tiny to huge
	// ═══════════════════════════════════════════
	{
		Name: "TinyLlama 1.1B", OllamaTag: "tinyllama:1.1b", Family: "llama",
		ParameterSize: "1.1B", ParameterCount: 1.1, QuantLevel: "Q4_K_M",
		DiskSizeGB: 0.6, EstimatedVRAMGB: 1.2,
		UseCases: []UseCase{UseCaseChat}, Quality: 12,
		Description: "Smallest Llama — Raspberry Pi & edge devices",
	},
	{
		Name: "SmolLM2 135M", OllamaTag: "smollm2:135m", Family: "smollm",
		ParameterSize: "135M", ParameterCount: 0.135, QuantLevel: "Q8_0",
		DiskSizeGB: 0.15, EstimatedVRAMGB: 0.3,
		UseCases: []UseCase{UseCaseChat}, Quality: 5,
		Description: "Ultra-tiny model for IoT and embedded hardware",
	},
	{
		Name: "SmolLM2 360M", OllamaTag: "smollm2:360m", Family: "smollm",
		ParameterSize: "360M", ParameterCount: 0.36, QuantLevel: "Q8_0",
		DiskSizeGB: 0.35, EstimatedVRAMGB: 0.6,
		UseCases: []UseCase{UseCaseChat}, Quality: 10,
		Description: "Tiny model for Raspberry Pi and low-RAM devices",
	},
	{
		Name: "SmolLM2 1.7B", OllamaTag: "smollm2:1.7b", Family: "smollm",
		ParameterSize: "1.7B", ParameterCount: 1.7, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.0, EstimatedVRAMGB: 2.0,
		UseCases: []UseCase{UseCaseChat}, Quality: 20,
		Description: "Tiny chat model for very constrained hardware",
	},
	{
		Name: "Llama 3.2 1B", OllamaTag: "llama3.2:1b", Family: "llama",
		ParameterSize: "1B", ParameterCount: 1.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 0.7, EstimatedVRAMGB: 1.5,
		UseCases: []UseCase{UseCaseChat}, Quality: 18,
		Description: "Meta's smallest Llama — edge and mobile",
	},
	{
		Name: "Llama 3.2 3B", OllamaTag: "llama3.2:3b", Family: "llama",
		ParameterSize: "3B", ParameterCount: 3.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 2.0, EstimatedVRAMGB: 3.5,
		UseCases: []UseCase{UseCaseChat}, Quality: 35,
		Description: "Small and fast chat model from Meta",
	},
	{
		Name: "Phi-4 Mini 3.8B", OllamaTag: "phi4-mini", Family: "phi",
		ParameterSize: "3.8B", ParameterCount: 3.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 2.4, EstimatedVRAMGB: 4.0,
		UseCases: []UseCase{UseCaseChat, UseCaseCode}, Quality: 42,
		Description: "Microsoft's efficient small model",
	},
	{
		Name: "Gemma 3 4B", OllamaTag: "gemma3:4b", Family: "gemma",
		ParameterSize: "4B", ParameterCount: 4.3, QuantLevel: "Q4_K_M",
		DiskSizeGB: 3.0, EstimatedVRAMGB: 4.5,
		UseCases: []UseCase{UseCaseChat, UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 45, Description: "Google's compact multimodal model",
	},
	{
		Name: "Mistral 7B", OllamaTag: "mistral:7b", Family: "mistral",
		ParameterSize: "7B", ParameterCount: 7.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.1, EstimatedVRAMGB: 6.0,
		UseCases: []UseCase{UseCaseChat}, Quality: 55,
		Description: "Strong general-purpose 7B model",
	},
	{
		Name: "Llama 3.1 8B", OllamaTag: "llama3.1:8b", Family: "llama",
		ParameterSize: "8B", ParameterCount: 8.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.7, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseChat}, Capabilities: []string{"tools"},
		Quality: 60, Description: "Meta's flagship 8B model with tool support",
	},
	{
		Name: "Qwen 3 8B", OllamaTag: "qwen3:8b", Family: "qwen",
		ParameterSize: "8B", ParameterCount: 8.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 5.0, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseChat, UseCaseReasoning}, Capabilities: []string{"tools", "thinking"},
		Quality: 64, Description: "Alibaba's latest 8B with hybrid thinking",
	},
	{
		Name: "Gemma 3 12B", OllamaTag: "gemma3:12b", Family: "gemma",
		ParameterSize: "12B", ParameterCount: 12.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 7.3, EstimatedVRAMGB: 10.0,
		UseCases: []UseCase{UseCaseChat, UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 70, Description: "Google's strong multimodal model",
	},
	{
		Name: "Qwen 2.5 14B", OllamaTag: "qwen2.5:14b", Family: "qwen",
		ParameterSize: "14B", ParameterCount: 14.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 9.0, EstimatedVRAMGB: 12.0,
		UseCases: []UseCase{UseCaseChat}, Capabilities: []string{"tools"},
		Quality: 72, Description: "Alibaba's versatile 14B model",
	},
	{
		Name: "Mistral Small 24B", OllamaTag: "mistral-small:24b", Family: "mistral",
		ParameterSize: "24B", ParameterCount: 24.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 14.0, EstimatedVRAMGB: 16.0,
		UseCases: []UseCase{UseCaseChat}, Capabilities: []string{"tools"},
		Quality: 78, Description: "Mistral's efficient mid-range model",
	},
	{
		Name: "Gemma 3 27B", OllamaTag: "gemma3:27b", Family: "gemma",
		ParameterSize: "27B", ParameterCount: 27.4, QuantLevel: "Q4_K_M",
		DiskSizeGB: 17.0, EstimatedVRAMGB: 20.0,
		UseCases: []UseCase{UseCaseChat, UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 82, Description: "Google's largest Gemma — excellent quality",
	},
	{
		Name: "Qwen 2.5 32B", OllamaTag: "qwen2.5:32b", Family: "qwen",
		ParameterSize: "32B", ParameterCount: 32.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 20.0, EstimatedVRAMGB: 22.0,
		UseCases: []UseCase{UseCaseChat}, Capabilities: []string{"tools"},
		Quality: 84, Description: "Alibaba's large model with excellent tool use",
	},
	{
		Name: "Llama 3.1 70B", OllamaTag: "llama3.1:70b", Family: "llama",
		ParameterSize: "70B", ParameterCount: 70.6, QuantLevel: "Q4_K_M",
		DiskSizeGB: 40.0, EstimatedVRAMGB: 44.0,
		UseCases: []UseCase{UseCaseChat}, Capabilities: []string{"tools"},
		Quality: 90, Description: "Meta's large model — near-frontier quality",
	},
	{
		Name: "Qwen 2.5 72B", OllamaTag: "qwen2.5:72b", Family: "qwen",
		ParameterSize: "72B", ParameterCount: 72.7, QuantLevel: "Q4_K_M",
		DiskSizeGB: 42.0, EstimatedVRAMGB: 48.0,
		UseCases: []UseCase{UseCaseChat}, Capabilities: []string{"tools"},
		Quality: 92, Description: "Alibaba's largest — top-tier quality",
	},

	// ═══════════════════════════════════════════
	// CODE
	// ═══════════════════════════════════════════
	{
		Name: "Qwen 2.5 Coder 0.5B", OllamaTag: "qwen2.5-coder:0.5b", Family: "qwen",
		ParameterSize: "0.5B", ParameterCount: 0.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 0.4, EstimatedVRAMGB: 1.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 15,
		Description: "Tiny code completion model",
	},
	{
		Name: "Qwen 2.5 Coder 1.5B", OllamaTag: "qwen2.5-coder:1.5b", Family: "qwen",
		ParameterSize: "1.5B", ParameterCount: 1.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.0, EstimatedVRAMGB: 2.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 28,
		Description: "Small but capable code model",
	},
	{
		Name: "StarCoder2 3B", OllamaTag: "starcoder2:3b", Family: "starcoder",
		ParameterSize: "3B", ParameterCount: 3.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.8, EstimatedVRAMGB: 3.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 32,
		Description: "BigCode's compact code model — strong FIM support",
	},
	{
		Name: "Qwen 2.5 Coder 3B", OllamaTag: "qwen2.5-coder:3b", Family: "qwen",
		ParameterSize: "3B", ParameterCount: 3.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.9, EstimatedVRAMGB: 3.5,
		UseCases: []UseCase{UseCaseCode}, Quality: 38,
		Description: "Balanced code model for constrained hardware",
	},
	{
		Name: "Granite Code 3B", OllamaTag: "granite-code:3b", Family: "granite",
		ParameterSize: "3B", ParameterCount: 3.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.9, EstimatedVRAMGB: 3.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 34,
		Description: "IBM's enterprise-focused code model",
	},
	{
		Name: "Yi Coder 9B", OllamaTag: "yi-coder:9b", Family: "yi",
		ParameterSize: "9B", ParameterCount: 8.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 5.0, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseCode}, Quality: 52,
		Description: "01.AI's code model — strong in multiple languages",
	},
	{
		Name: "CodeGemma 7B", OllamaTag: "codegemma:7b", Family: "gemma",
		ParameterSize: "7B", ParameterCount: 8.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 5.0, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseCode}, Quality: 50,
		Description: "Google's code-focused model",
	},
	{
		Name: "Qwen 2.5 Coder 7B", OllamaTag: "qwen2.5-coder:7b", Family: "qwen",
		ParameterSize: "7B", ParameterCount: 7.6, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.7, EstimatedVRAMGB: 6.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 58,
		Description: "Strong code model for mid-range hardware",
	},
	{
		Name: "StarCoder2 15B", OllamaTag: "starcoder2:15b", Family: "starcoder",
		ParameterSize: "15B", ParameterCount: 15.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 9.0, EstimatedVRAMGB: 11.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 64,
		Description: "BigCode's large model — excellent multi-language support",
	},
	{
		Name: "DeepSeek Coder V2 16B", OllamaTag: "deepseek-coder-v2:16b", Family: "deepseek",
		ParameterSize: "16B", ParameterCount: 15.7, QuantLevel: "Q4_K_M",
		DiskSizeGB: 8.9, EstimatedVRAMGB: 11.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 68,
		Description: "DeepSeek's MoE code model — efficient and strong",
	},
	{
		Name: "Qwen 2.5 Coder 14B", OllamaTag: "qwen2.5-coder:14b", Family: "qwen",
		ParameterSize: "14B", ParameterCount: 14.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 9.0, EstimatedVRAMGB: 12.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 72,
		Description: "Excellent code model for high-end hardware",
	},
	{
		Name: "Codestral 22B", OllamaTag: "codestral:22b", Family: "mistral",
		ParameterSize: "22B", ParameterCount: 22.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 13.0, EstimatedVRAMGB: 16.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 76,
		Description: "Mistral's dedicated code model",
	},
	{
		Name: "Qwen 2.5 Coder 32B", OllamaTag: "qwen2.5-coder:32b", Family: "qwen",
		ParameterSize: "32B", ParameterCount: 32.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 20.0, EstimatedVRAMGB: 22.0,
		UseCases: []UseCase{UseCaseCode}, Quality: 88,
		Description: "Top-tier open code model",
	},

	// ═══════════════════════════════════════════
	// VISION (image understanding)
	// ═══════════════════════════════════════════
	{
		Name: "Moondream 1.8B", OllamaTag: "moondream:1.8b", Family: "moondream",
		ParameterSize: "1.8B", ParameterCount: 1.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.1, EstimatedVRAMGB: 2.5,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 25, Description: "Tiny vision model — basic image understanding",
	},
	{
		Name: "LLaVA 7B", OllamaTag: "llava:7b", Family: "llava",
		ParameterSize: "7B", ParameterCount: 7.1, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.5, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 50, Description: "Classic vision-language model",
	},
	{
		Name: "MiniCPM-V 8B", OllamaTag: "minicpm-v:8b", Family: "minicpm",
		ParameterSize: "8B", ParameterCount: 8.1, QuantLevel: "Q4_K_M",
		DiskSizeGB: 5.5, EstimatedVRAMGB: 7.5,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 55, Description: "Compact yet capable vision model",
	},
	{
		Name: "Qwen 2.5 VL 7B", OllamaTag: "qwen2.5vl:7b", Family: "qwen",
		ParameterSize: "7B", ParameterCount: 8.3, QuantLevel: "Q4_K_M",
		DiskSizeGB: 5.4, EstimatedVRAMGB: 8.0,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 62, Description: "Alibaba's vision-language model",
	},
	{
		Name: "Llama 3.2 Vision 11B", OllamaTag: "llama3.2-vision:11b", Family: "llama",
		ParameterSize: "11B", ParameterCount: 11.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 7.9, EstimatedVRAMGB: 10.0,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 65, Description: "Meta's vision model — strong image understanding",
	},
	{
		Name: "InternVL2 8B", OllamaTag: "internvl2:8b", Family: "internvl",
		ParameterSize: "8B", ParameterCount: 8.1, QuantLevel: "Q4_K_M",
		DiskSizeGB: 5.2, EstimatedVRAMGB: 7.0,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 58, Description: "Shanghai AI Lab's strong vision-language model",
	},
	{
		Name: "Pixtral 12B", OllamaTag: "pixtral:12b", Family: "mistral",
		ParameterSize: "12B", ParameterCount: 12.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 7.5, EstimatedVRAMGB: 10.0,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 68, Description: "Mistral's native multimodal model",
	},
	{
		Name: "Molmo 7B", OllamaTag: "molmo:7b", Family: "molmo",
		ParameterSize: "7B", ParameterCount: 7.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.5, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 56, Description: "Allen AI's open vision model with pointing",
	},
	{
		Name: "Qwen 2.5 VL 32B", OllamaTag: "qwen2.5vl:32b", Family: "qwen",
		ParameterSize: "32B", ParameterCount: 32.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 20.0, EstimatedVRAMGB: 24.0,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 82, Description: "Alibaba's largest vision model — near-frontier",
	},
	{
		Name: "Llama 3.2 Vision 90B", OllamaTag: "llama3.2-vision:90b", Family: "llama",
		ParameterSize: "90B", ParameterCount: 90.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 52.0, EstimatedVRAMGB: 58.0,
		UseCases: []UseCase{UseCaseVision}, Capabilities: []string{"vision"},
		Quality: 90, Description: "Meta's largest vision model — top quality",
	},

	// ═══════════════════════════════════════════
	// EMBEDDING
	// ═══════════════════════════════════════════
	{
		Name: "All-MiniLM L6", OllamaTag: "all-minilm:l6-v2", Family: "minilm",
		ParameterSize: "22M", ParameterCount: 0.022, QuantLevel: "F32",
		DiskSizeGB: 0.05, EstimatedVRAMGB: 0.2,
		UseCases: []UseCase{UseCaseEmbedding}, Capabilities: []string{"embedding"},
		Quality: 30, Description: "Tiny embedding model — fast but basic quality",
	},
	{
		Name: "Nomic Embed Text", OllamaTag: "nomic-embed-text", Family: "nomic",
		ParameterSize: "137M", ParameterCount: 0.137, QuantLevel: "F16",
		DiskSizeGB: 0.27, EstimatedVRAMGB: 0.5,
		UseCases: []UseCase{UseCaseEmbedding}, Capabilities: []string{"embedding"},
		Quality: 55, Description: "Most popular embedding model on Ollama",
	},
	{
		Name: "MxBai Embed Large", OllamaTag: "mxbai-embed-large", Family: "mxbai",
		ParameterSize: "335M", ParameterCount: 0.335, QuantLevel: "F16",
		DiskSizeGB: 0.67, EstimatedVRAMGB: 1.0,
		UseCases: []UseCase{UseCaseEmbedding}, Capabilities: []string{"embedding"},
		Quality: 70, Description: "High-quality embedding model",
	},
	{
		Name: "BGE-M3", OllamaTag: "bge-m3", Family: "bge",
		ParameterSize: "567M", ParameterCount: 0.567, QuantLevel: "F16",
		DiskSizeGB: 1.2, EstimatedVRAMGB: 1.5,
		UseCases: []UseCase{UseCaseEmbedding}, Capabilities: []string{"embedding"},
		Quality: 80, Description: "Multilingual embedding model — top quality",
	},
	{
		Name: "Snowflake Arctic Embed L", OllamaTag: "snowflake-arctic-embed:335m", Family: "snowflake",
		ParameterSize: "335M", ParameterCount: 0.335, QuantLevel: "F16",
		DiskSizeGB: 0.67, EstimatedVRAMGB: 1.0,
		UseCases: []UseCase{UseCaseEmbedding}, Capabilities: []string{"embedding"},
		Quality: 72, Description: "Snowflake's efficient retrieval-focused embeddings",
	},

	// ═══════════════════════════════════════════
	// REASONING (chain-of-thought)
	// ═══════════════════════════════════════════
	{
		Name: "DeepSeek R1 1.5B", OllamaTag: "deepseek-r1:1.5b", Family: "deepseek",
		ParameterSize: "1.5B", ParameterCount: 1.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.0, EstimatedVRAMGB: 2.0,
		UseCases: []UseCase{UseCaseReasoning}, Capabilities: []string{"thinking"},
		Quality: 22, Description: "Tiny reasoning model with chain-of-thought",
	},
	{
		Name: "DeepSeek R1 7B", OllamaTag: "deepseek-r1:7b", Family: "deepseek",
		ParameterSize: "7B", ParameterCount: 7.6, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.7, EstimatedVRAMGB: 6.0,
		UseCases: []UseCase{UseCaseReasoning, UseCaseChat}, Capabilities: []string{"thinking"},
		Quality: 52, Description: "Reasoning model — good balance of size and capability",
	},
	{
		Name: "DeepSeek R1 14B", OllamaTag: "deepseek-r1:14b", Family: "deepseek",
		ParameterSize: "14B", ParameterCount: 14.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 9.0, EstimatedVRAMGB: 12.0,
		UseCases: []UseCase{UseCaseReasoning, UseCaseChat}, Capabilities: []string{"thinking"},
		Quality: 68, Description: "Strong reasoning model with explicit chain-of-thought",
	},
	{

		Name: "Phi-4 Reasoning 14B", OllamaTag: "phi4-reasoning:14b", Family: "phi",
		ParameterSize: "14B", ParameterCount: 14.7, QuantLevel: "Q4_K_M",
		DiskSizeGB: 9.0, EstimatedVRAMGB: 12.0,
		UseCases: []UseCase{UseCaseReasoning}, Capabilities: []string{"thinking"},
		Quality: 70, Description: "Microsoft's reasoning-tuned Phi-4",
	},
	{
		Name: "DeepSeek R1 32B", OllamaTag: "deepseek-r1:32b", Family: "deepseek",
		ParameterSize: "32B", ParameterCount: 32.8, QuantLevel: "Q4_K_M",
		DiskSizeGB: 20.0, EstimatedVRAMGB: 22.0,
		UseCases: []UseCase{UseCaseReasoning, UseCaseChat}, Capabilities: []string{"thinking"},
		Quality: 82, Description: "Powerful reasoning — excellent for complex problems",
	},
	{
		Name: "QwQ 32B", OllamaTag: "qwq:32b", Family: "qwen",
		ParameterSize: "32B", ParameterCount: 32.5, QuantLevel: "Q4_K_M",
		DiskSizeGB: 20.0, EstimatedVRAMGB: 22.0,
		UseCases: []UseCase{UseCaseReasoning}, Capabilities: []string{"thinking"},
		Quality: 84, Description: "Alibaba's dedicated reasoning model — strong on math & logic",
	},
	{
		Name: "DeepSeek R1 70B", OllamaTag: "deepseek-r1:70b", Family: "deepseek",
		ParameterSize: "70B", ParameterCount: 70.6, QuantLevel: "Q4_K_M",
		DiskSizeGB: 40.0, EstimatedVRAMGB: 44.0,
		UseCases: []UseCase{UseCaseReasoning, UseCaseChat}, Capabilities: []string{"thinking"},
		Quality: 94, Description: "Top-tier reasoning — near-frontier intelligence",
	},

	// ═══════════════════════════════════════════
	// IMAGE GENERATION
	// Note: These models use diffusion pipelines (ComfyUI, A1111).
	// They are NOT Ollama chat models — use IsExternal + Pipeline fields.
	// ═══════════════════════════════════════════
	{
		Name: "Stable Diffusion 3.5 Large Turbo", OllamaTag: "sd3.5-large-turbo", Family: "sd3",
		ParameterSize: "8B", ParameterCount: 8.0, QuantLevel: "F16",
		DiskSizeGB: 16.0, EstimatedVRAMGB: 12.0,
		UseCases: []UseCase{UseCaseImage}, Capabilities: []string{"image-generation"},
		Quality: 72, Description: "Stability AI's fast image generation — 4 steps",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/stabilityai/stable-diffusion-3.5-large-turbo",
	},
	{
		Name: "Stable Diffusion 3.5 Large", OllamaTag: "sd3.5-large", Family: "sd3",
		ParameterSize: "8B", ParameterCount: 8.0, QuantLevel: "F16",
		DiskSizeGB: 16.0, EstimatedVRAMGB: 16.0,
		UseCases: []UseCase{UseCaseImage}, Capabilities: []string{"image-generation"},
		Quality: 78, Description: "Stability AI's high-quality image generation",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/stabilityai/stable-diffusion-3.5-large",
	},
	{
		Name: "FLUX.1 Schnell", OllamaTag: "flux-schnell", Family: "flux",
		ParameterSize: "12B", ParameterCount: 12.0, QuantLevel: "F16",
		DiskSizeGB: 24.0, EstimatedVRAMGB: 16.0,
		UseCases: []UseCase{UseCaseImage}, Capabilities: []string{"image-generation"},
		Quality: 80, Description: "Black Forest Labs' fast image generation — 4 steps",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/black-forest-labs/FLUX.1-schnell",
	},
	{
		Name: "FLUX.1 Dev", OllamaTag: "flux-dev", Family: "flux",
		ParameterSize: "12B", ParameterCount: 12.0, QuantLevel: "F16",
		DiskSizeGB: 24.0, EstimatedVRAMGB: 24.0,
		UseCases: []UseCase{UseCaseImage}, Capabilities: []string{"image-generation"},
		Quality: 88, Description: "Black Forest Labs' high-quality image generation — 30+ steps",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/black-forest-labs/FLUX.1-dev",
	},
	{
		Name: "PixArt Sigma XL 2", OllamaTag: "pixart-sigma", Family: "pixart",
		ParameterSize: "0.6B", ParameterCount: 0.6, QuantLevel: "F16",
		DiskSizeGB: 2.5, EstimatedVRAMGB: 6.0,
		UseCases: []UseCase{UseCaseImage}, Capabilities: []string{"image-generation"},
		Quality: 55, Description: "Lightweight diffusion transformer — fast on low-VRAM GPUs",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/PixArt-alpha/PixArt-Sigma",
	},
	{
		Name: "SDXL 1.0", OllamaTag: "sdxl", Family: "sdxl",
		ParameterSize: "3.5B", ParameterCount: 3.5, QuantLevel: "F16",
		DiskSizeGB: 7.0, EstimatedVRAMGB: 8.0,
		UseCases: []UseCase{UseCaseImage}, Capabilities: []string{"image-generation"},
		Quality: 65, Description: "Classic high-resolution image generation — huge LoRA ecosystem",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/stabilityai/stable-diffusion-xl-base-1.0",
	},

	// ═══════════════════════════════════════════
	// VIDEO GENERATION
	// Note: These models use diffusion pipelines (ComfyUI).
	// They are NOT Ollama chat models — use IsExternal + Pipeline fields.
	// ═══════════════════════════════════════════
	{
		Name: "Wan 2.1 1.3B", OllamaTag: "wan2.1:1.3b", Family: "wan",
		ParameterSize: "1.3B", ParameterCount: 1.3, QuantLevel: "F16",
		DiskSizeGB: 2.8, EstimatedVRAMGB: 4.0,
		UseCases: []UseCase{UseCaseVideo}, Capabilities: []string{"video-generation"},
		Quality: 30, Description: "Tiny text-to-video model — run via ComfyUI",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/Wan-AI/Wan2.1-T2V-1.3B",
	},
	{
		Name: "LTX Video 0.9B", OllamaTag: "ltx-video", Family: "ltx",
		ParameterSize: "0.9B", ParameterCount: 0.9, QuantLevel: "F16",
		DiskSizeGB: 2.0, EstimatedVRAMGB: 6.0,
		UseCases: []UseCase{UseCaseVideo}, Capabilities: []string{"video-generation"},
		Quality: 40, Description: "Lightricks' fast text-to-video — real-time capable",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/Lightricks/LTX-Video",
	},
	{
		Name: "CogVideoX 5B", OllamaTag: "cogvideox:5b", Family: "cogvideo",
		ParameterSize: "5B", ParameterCount: 5.0, QuantLevel: "F16",
		DiskSizeGB: 10.0, EstimatedVRAMGB: 12.0,
		UseCases: []UseCase{UseCaseVideo}, Capabilities: []string{"video-generation"},
		Quality: 55, Description: "Tsinghua's diffusion transformer for video generation",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/THUDM/CogVideoX-5b",
	},
	{
		Name: "HunyuanVideo 13B", OllamaTag: "hunyuan-video", Family: "hunyuan",
		ParameterSize: "13B", ParameterCount: 13.0, QuantLevel: "F16",
		DiskSizeGB: 26.0, EstimatedVRAMGB: 24.0,
		UseCases: []UseCase{UseCaseVideo}, Capabilities: []string{"video-generation"},
		Quality: 70, Description: "Tencent's high-quality text-to-video model",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/tencent/HunyuanVideo",
	},
	{
		Name: "Wan 2.1 14B", OllamaTag: "wan2.1:14b", Family: "wan",
		ParameterSize: "14B", ParameterCount: 14.0, QuantLevel: "F16",
		DiskSizeGB: 28.0, EstimatedVRAMGB: 32.0,
		UseCases: []UseCase{UseCaseVideo}, Capabilities: []string{"video-generation"},
		Quality: 75, Description: "Full text-to-video model — run via ComfyUI",
		IsExternal: true, Pipeline: "comfyui",
		ExternalURL: "https://huggingface.co/Wan-AI/Wan2.1-T2V-14B",
	},

	// ═══════════════════════════════════════════
	// AUDIO (speech-to-text, text-to-speech)
	// Note: Some models run via Ollama, others need
	// dedicated pipelines (whisper.cpp, Piper, etc.)
	// ═══════════════════════════════════════════
	{
		Name: "Whisper Tiny", OllamaTag: "whisper:tiny", Family: "whisper",
		ParameterSize: "39M", ParameterCount: 0.039, QuantLevel: "F16",
		DiskSizeGB: 0.08, EstimatedVRAMGB: 0.3,
		UseCases: []UseCase{UseCaseAudio}, Capabilities: []string{"speech-to-text"},
		Quality: 20, Description: "OpenAI's tiniest speech recognition — Raspberry Pi capable",
	},
	{
		Name: "Whisper Base", OllamaTag: "whisper:base", Family: "whisper",
		ParameterSize: "74M", ParameterCount: 0.074, QuantLevel: "F16",
		DiskSizeGB: 0.15, EstimatedVRAMGB: 0.5,
		UseCases: []UseCase{UseCaseAudio}, Capabilities: []string{"speech-to-text"},
		Quality: 35, Description: "Good speech recognition for constrained hardware",
	},
	{
		Name: "Whisper Small", OllamaTag: "whisper:small", Family: "whisper",
		ParameterSize: "244M", ParameterCount: 0.244, QuantLevel: "F16",
		DiskSizeGB: 0.5, EstimatedVRAMGB: 1.0,
		UseCases: []UseCase{UseCaseAudio}, Capabilities: []string{"speech-to-text"},
		Quality: 50, Description: "Balanced speech recognition — good accuracy",
	},
	{
		Name: "Whisper Medium", OllamaTag: "whisper:medium", Family: "whisper",
		ParameterSize: "769M", ParameterCount: 0.769, QuantLevel: "F16",
		DiskSizeGB: 1.5, EstimatedVRAMGB: 2.5,
		UseCases: []UseCase{UseCaseAudio}, Capabilities: []string{"speech-to-text"},
		Quality: 65, Description: "Strong multilingual speech recognition",
	},
	{
		Name: "Whisper Large V3", OllamaTag: "whisper:large-v3", Family: "whisper",
		ParameterSize: "1.5B", ParameterCount: 1.55, QuantLevel: "F16",
		DiskSizeGB: 3.1, EstimatedVRAMGB: 4.0,
		UseCases: []UseCase{UseCaseAudio}, Capabilities: []string{"speech-to-text"},
		Quality: 82, Description: "Best open speech recognition — 99+ languages",
	},
	{
		Name: "Bark", OllamaTag: "bark", Family: "bark",
		ParameterSize: "0.4B", ParameterCount: 0.4, QuantLevel: "F16",
		DiskSizeGB: 1.5, EstimatedVRAMGB: 4.0,
		UseCases: []UseCase{UseCaseAudio}, Capabilities: []string{"text-to-speech"},
		Quality: 45, Description: "Suno's text-to-speech with voice cloning — music & sound effects",
	},

	// ═══════════════════════════════════════════
	// Unrestricted / UNCENSORED
	// Models without built-in content filters.
	// For image/video Unrestricted generation, use Stable Diffusion
	// (ComfyUI, A1111, Forge) with appropriate checkpoints.
	// ═══════════════════════════════════════════
	{
		Name: "Dolphin Phi 2.7B", OllamaTag: "dolphin-phi:2.7b", Family: "phi",
		ParameterSize: "2.7B", ParameterCount: 2.7, QuantLevel: "Q4_K_M",
		DiskSizeGB: 1.6, EstimatedVRAMGB: 3.0,
		UseCases: []UseCase{UseCaseUnrestricted, UseCaseChat}, Capabilities: []string{"unrestricted"},
		Quality: 28, Description: "Tiny unrestricted chat model — very low hardware requirements",
	},
	{
		Name: "Llama2 Uncensored 7B", OllamaTag: "llama2-uncensored:7b", Family: "llama",
		ParameterSize: "7B", ParameterCount: 7.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 3.8, EstimatedVRAMGB: 5.5,
		UseCases: []UseCase{UseCaseUnrestricted, UseCaseChat}, Capabilities: []string{"unrestricted"},
		Quality: 40, Description: "Unrestricted Llama 2 — no built-in content filters",
	},
	{
		Name: "Dolphin Mistral 7B", OllamaTag: "dolphin-mistral:7b", Family: "mistral",
		ParameterSize: "7B", ParameterCount: 7.2, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.1, EstimatedVRAMGB: 6.0,
		UseCases: []UseCase{UseCaseUnrestricted, UseCaseChat}, Capabilities: []string{"unrestricted"},
		Quality: 52, Description: "Unrestricted Mistral — strong general-purpose without filters",
	},
	{
		Name: "Dolphin Llama3 8B", OllamaTag: "dolphin-llama3:8b", Family: "llama",
		ParameterSize: "8B", ParameterCount: 8.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 4.7, EstimatedVRAMGB: 6.5,
		UseCases: []UseCase{UseCaseUnrestricted, UseCaseChat}, Capabilities: []string{"unrestricted"},
		Quality: 58, Description: "Unrestricted Llama 3 — high quality without content restrictions",
	},
	{
		Name: "Wizard Vicuna Uncensored 13B", OllamaTag: "wizard-vicuna-uncensored:13b", Family: "llama",
		ParameterSize: "13B", ParameterCount: 13.0, QuantLevel: "Q4_K_M",
		DiskSizeGB: 7.4, EstimatedVRAMGB: 10.0,
		UseCases: []UseCase{UseCaseUnrestricted, UseCaseChat}, Capabilities: []string{"unrestricted"},
		Quality: 65, Description: "Classic unrestricted model — well-known for open output",
	},
	{
		Name: "Dolphin Mixtral 8x7B", OllamaTag: "dolphin-mixtral:8x7b", Family: "mixtral",
		ParameterSize: "47B (MoE)", ParameterCount: 46.7, QuantLevel: "Q4_K_M",
		DiskSizeGB: 26.0, EstimatedVRAMGB: 28.0,
		UseCases: []UseCase{UseCaseUnrestricted, UseCaseChat}, Capabilities: []string{"unrestricted"},
		Quality: 78, Description: "Unrestricted Mixtral MoE — top quality open model",
	},
}

// GetModelsByUseCase returns all models matching the given use case.
func GetModelsByUseCase(useCase UseCase) []ModelEntry {
	var result []ModelEntry
	for _, m := range Catalog {
		for _, uc := range m.UseCases {
			if uc == useCase {
				result = append(result, m)
				break
			}
		}
	}
	return result
}

// GetModelByTag finds a model by its Ollama tag.
func GetModelByTag(tag string) *ModelEntry {
	for i, m := range Catalog {
		if m.OllamaTag == tag {
			return &Catalog[i]
		}
	}
	return nil
}
