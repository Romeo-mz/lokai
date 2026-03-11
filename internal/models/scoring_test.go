package models

import "testing"

// ─── inferQuantLevel ─────────────────────────────────────────────────────────

func TestInferQuantLevel_KnownLevels(t *testing.T) {
	tests := []struct {
		tag      string
		expected string
	}{
		{"llama3:8b-q4_k_m", "Q4_K_M"},
		{"llama3:8b-Q4_0", "Q4_0"},
		{"llama3:70b-q8_0", "Q8_0"},
		{"llama3:8b-fp16", "F16"},
		{"llama3:8b-F16", "F16"},
		{"codellama:7b-q6_k", "Q6_K"},
		{"mistral:7b-q5_k_m", "Q5_K_M"},
		{"phi3:3.8b-q3_k_m", "Q3_K_M"},
		{"llama3:8b-q2_k", "Q2_K"},
	}
	for _, tt := range tests {
		got := inferQuantLevel(tt.tag)
		if got != tt.expected {
			t.Errorf("inferQuantLevel(%q) = %q, want %q", tt.tag, got, tt.expected)
		}
	}
}

func TestInferQuantLevel_UnknownDefaultsToQ4KM(t *testing.T) {
	tests := []string{
		"llama3:latest",
		"llama3:8b",
		"smollm2:135m",
		"unknown-model:some-tag",
	}
	for _, tag := range tests {
		got := inferQuantLevel(tag)
		if got != "Q4_K_M" {
			t.Errorf("inferQuantLevel(%q) = %q, want Q4_K_M (default)", tag, got)
		}
	}
}

func TestQuantVRAMMultiplier_Q8ScalesUpFromQ4(t *testing.T) {
	q4Mult := quantVRAMMultiplier["Q4_K_M"]
	q8Mult := quantVRAMMultiplier["Q8_0"]
	if q8Mult <= q4Mult {
		t.Errorf("Q8_0 multiplier (%.2f) should be larger than Q4_K_M (%.2f)", q8Mult, q4Mult)
	}
}

func TestQuantVRAMMultiplier_F16IsLargest(t *testing.T) {
	f16 := quantVRAMMultiplier["F16"]
	for level, mult := range quantVRAMMultiplier {
		if level == "F32" {
			continue // F32 is larger than F16 intentionally
		}
		if mult > f16 {
			t.Errorf("F16 multiplier (%.2f) should be the second largest, but %s (%.2f) is larger", f16, level, mult)
		}
	}
}

// ─── subTaskBonus ─────────────────────────────────────────────────────────────

func modelWithCap(caps ...string) ModelEntry {
	return ModelEntry{Capabilities: caps}
}

func modelWithFamily(family string) ModelEntry {
	return ModelEntry{Family: family}
}

func modelWithTag(tag string) ModelEntry {
	return ModelEntry{OllamaTag: tag}
}

func modelWithParams(params float64) ModelEntry {
	return ModelEntry{ParameterCount: params}
}

func TestSubTaskBonus_NoSubTask(t *testing.T) {
	bonus := subTaskBonus(ModelEntry{}, "")
	if bonus != 0 {
		t.Errorf("empty sub-task should return 0, got %.1f", bonus)
	}
}

func TestSubTaskBonus_TranscriptionUsesCapability(t *testing.T) {
	// A model with the speech-to-text capability should receive the bonus.
	m := modelWithCap("speech-to-text")
	bonus := subTaskBonus(m, SubTaskTranscription)
	if bonus != 15 {
		t.Errorf("SubTaskTranscription with speech-to-text cap: expected 15, got %.1f", bonus)
	}

	// A model without the capability should not.
	noCapModel := ModelEntry{}
	bonus = subTaskBonus(noCapModel, SubTaskTranscription)
	if bonus != 0 {
		t.Errorf("SubTaskTranscription without cap: expected 0, got %.1f", bonus)
	}
}

func TestSubTaskBonus_TTSUsesCapability(t *testing.T) {
	m := modelWithCap("text-to-speech")
	bonus := subTaskBonus(m, SubTaskTTS)
	if bonus != 15 {
		t.Errorf("SubTaskTTS with text-to-speech cap: expected 15, got %.1f", bonus)
	}
}

func TestSubTaskBonus_PhotorealisticUsesFamily(t *testing.T) {
	for _, family := range []string{"flux", "sd3"} {
		m := modelWithFamily(family)
		bonus := subTaskBonus(m, SubTaskPhotorealistic)
		if bonus != 15 {
			t.Errorf("SubTaskPhotorealistic family=%q: expected 15, got %.1f", family, bonus)
		}
	}

	// Non-matching family.
	m := modelWithFamily("sdxl")
	bonus := subTaskBonus(m, SubTaskPhotorealistic)
	if bonus != 0 {
		t.Errorf("SubTaskPhotorealistic family=sdxl: expected 0, got %.1f", bonus)
	}
}

func TestSubTaskBonus_ArtisticUsesFamily(t *testing.T) {
	for _, family := range []string{"sdxl", "pixart"} {
		m := modelWithFamily(family)
		bonus := subTaskBonus(m, SubTaskArtistic)
		if bonus != 12 {
			t.Errorf("SubTaskArtistic family=%q: expected 12, got %.1f", family, bonus)
		}
	}

	// Flux gets a lower bonus.
	m := modelWithFamily("flux")
	bonus := subTaskBonus(m, SubTaskArtistic)
	if bonus != 8 {
		t.Errorf("SubTaskArtistic family=flux: expected 8, got %.1f", bonus)
	}
}

func TestSubTaskBonus_FastDraftsUsesOllamaTag(t *testing.T) {
	m := modelWithTag("flux:schnell")
	bonus := subTaskBonus(m, SubTaskFastDrafts)
	if bonus != 15 {
		t.Errorf("SubTaskFastDrafts schnell tag: expected 15, got %.1f", bonus)
	}
}

func TestSubTaskBonus_MathUsesThinkingCapability(t *testing.T) {
	m := modelWithCap("thinking")
	bonus := subTaskBonus(m, SubTaskMath)
	if bonus != 15 {
		t.Errorf("SubTaskMath with thinking cap: expected 15, got %.1f", bonus)
	}
}

func TestSubTaskBonus_QuickQAFavorsSmallModels(t *testing.T) {
	small := modelWithParams(2.0)
	if bonus := subTaskBonus(small, SubTaskQuickQA); bonus != 15 {
		t.Errorf("QuickQA ≤3B: expected 15, got %.1f", bonus)
	}

	large := modelWithParams(70.0)
	bonus := subTaskBonus(large, SubTaskQuickQA)
	if bonus >= 0 {
		t.Errorf("QuickQA >30B should be penalised, got %.1f", bonus)
	}
}
