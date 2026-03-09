<div align="center">

# 🤖 lokai

**Find the best local AI model for your hardware — automatically**

[![CI](https://github.com/romeo-mz/lokai/actions/workflows/ci.yml/badge.svg)](https://github.com/romeo-mz/lokai/actions/workflows/ci.yml)
[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/romeo-mz/lokai?style=flat)](https://github.com/romeo-mz/lokai/releases)
[![Docker](https://img.shields.io/badge/Docker-ghcr.io-blue?logo=docker)](https://github.com/romeo-mz/lokai/pkgs/container/lokai)

Scan your hardware → Pick a use case → Get the best local AI model.
No guesswork. No VRAM spreadsheets. Just run `lokai`.

**76 models • 9 use cases • Every GPU from Raspberry Pi to workstation**

</div>

---

## Why lokai?

> *"Which model should I run on my GPU?"* is the #1 question in every local AI community. lokai answers it in 10 seconds.

| Without lokai | With lokai |
|---|---|
| Google "best 12GB VRAM model 2025" | Run `lokai` |
| Read 14 Reddit threads | Pick your use case |
| Compare model sizes vs VRAM | Get a ranked recommendation |
| Hope the model actually fits | Know it fits — with performance estimate |

## What it does

1. **Scans your hardware** — CPU, RAM, GPU (NVIDIA, AMD, Intel, Apple Silicon), VRAM
2. **Asks what you need** — Chat, Code, Vision, Embedding, Reasoning, Image Gen, Video, Audio, or Uncensored
3. **Recommends the best model** — ranked by quality, filtered by your VRAM budget
4. **Estimates performance** — tokens/sec, time-to-first-token, generation time
5. **Installs it for you** — pulls via Ollama in one click, with progress bar

## Quick Start

### Install

```bash
# With Go
go install github.com/romeo-mz/lokai/cmd/lokai@latest

# Or download binary from releases
# https://github.com/romeo-mz/lokai/releases
```

### Prerequisites

- [Ollama](https://ollama.com/) installed and running (`ollama serve`)

### Run

```bash
lokai
```

That's it. The interactive TUI guides you through everything.

### Docker

```bash
# With docker-compose (includes Ollama)
docker compose up -d ollama
docker compose run --rm lokai

# Standalone (connect to existing Ollama)
docker run --rm -it \
  -e OLLAMA_HOST=http://host.docker.internal:11434 \
  ghcr.io/romeo-mz/lokai:latest
```

## Commands & Flags

```bash
lokai                    # Interactive mode (default)
lokai --scan-only        # Just show hardware specs
lokai --benchmark        # Benchmark all installed models
lokai --clean            # Remove all installed models
lokai --version          # Print version
```

### Non-Interactive Mode

```bash
# JSON output for scripting
lokai --json --use-case code --priority balanced

# Specific use case
lokai --use-case chat --priority quality
```

| Flag | Description |
|------|-------------|
| `--json` | Output as JSON (non-interactive) |
| `--scan-only` | Only scan and display hardware specs |
| `--benchmark` | Benchmark all installed models |
| `--clean` | Remove all installed Ollama models |
| `--use-case` | Preset: `chat`, `code`, `vision`, `embedding`, `reasoning`, `image`, `video`, `audio`, `nsfw` |
| `--priority` | Preset: `quality`, `speed`, `balanced` |
| `--version` | Print version |

## Supported Hardware

| Platform | CPU | GPU | VRAM Detection |
|----------|-----|-----|----------------|
| **Linux** | ✅ Full | ✅ NVIDIA, AMD, Intel | ✅ NVML, sysfs |
| **macOS** | ✅ Full | ✅ Apple Silicon | ✅ Unified memory |
| **Windows** | ✅ Full | ✅ NVIDIA | ✅ NVML |
| **Raspberry Pi** | ✅ ARM64 | CPU-only | ✅ RAM-based budget |

### VRAM Budget Rules

| Hardware | Budget Calculation |
|----------|-------------------|
| NVIDIA GPU | Free VRAM × 90% |
| AMD GPU (Linux) | Free VRAM × 90% (sysfs) |
| Apple Silicon | Total RAM − 4 GB (OS overhead) |
| CPU-only | Available RAM × 70% |

## Model Catalog

**76 models** across 9 use cases — from 135M parameters (IoT) to 90B (workstation).

| Use Case | Models | Top Picks |
|----------|--------|-----------|
| 💬 **Chat** | 18 | Qwen 3, Gemma 3, Llama 3.1/3.2, Mistral |
| 💻 **Code** | 13 | Qwen 2.5 Coder, Codestral, StarCoder2, DeepSeek Coder |
| 👁 **Vision** | 10 | Qwen 2.5 VL, Pixtral, Llama 3.2 Vision, InternVL2 |
| 📐 **Embedding** | 5 | BGE-M3, MxBai, Nomic Embed, Snowflake Arctic |
| 🧠 **Reasoning** | 8 | DeepSeek R1, QwQ, Phi-4 Reasoning, Qwen 3 |
| 🖼 **Image Gen** | 6 | FLUX.1, SD 3.5, SDXL, PixArt |
| 🎬 **Video** | 5 | Wan 2.1, HunyuanVideo, CogVideoX, LTX Video |
| 🎙 **Audio** | 6 | Whisper (tiny→large), Bark |
| 🔓 **Uncensored** | 6 | Dolphin Llama3/Mistral/Mixtral, Wizard Vicuna |

### Hardware Tiers

| Tier | VRAM | Sweet Spot Models |
|------|------|-------------------|
| 🟢 RPi / Edge | < 2 GB | SmolLM2 135M, TinyLlama 1.1B |
| 🟡 Entry | 4-6 GB | Llama 3.2 3B, Phi-4 Mini, Gemma 3 4B |
| 🟠 Mid-range | 8-12 GB | Llama 3.1 8B, Qwen 2.5 Coder 7B, DeepSeek R1 7B |
| 🔴 High-end | 16-24 GB | Codestral 22B, Gemma 3 27B, DeepSeek R1 32B |
| 🟣 Workstation | 48+ GB | Llama 3.1 70B, Qwen 2.5 72B, DeepSeek R1 70B |

## How It Works

```
┌──────────────┐     ┌──────────────┐     ┌────────────────┐     ┌──────────────┐
│  Hardware    │───► │  Use Case    │────►│  Recommend     │────►│  Install /   │
│  Detection   │     │  Selection   │     │  Engine        │     │  Benchmark   │
│              │     │              │     │                │     │              │
│  • CPU/RAM   │     │  • Chat      │     │  • Filter by   │     │  ollama pull │
│  • GPU/VRAM  │     │  • Code      │     │    VRAM budget │     │  + progress  │
│  • Platform  │     │  • Vision    │     │  • Rank by     │     │  • tok/sec   │
│  • Features  │     │  • Embedding │     │    quality     │     │  • TTFT      │
│              │     │  • Reasoning │     │  • Estimate    │     │              │
│              │     │  • Image Gen │     │    performance │     │              │
│              │     │  • Video     │     │                │     │              │
│              │     │  • Audio     │     │                │     │              │
│              │     │  • NSFW      │     │                │     │              │
└──────────────┘     └──────────────┘     └────────────────┘     └──────────────┘
```

## Benchmarking

Benchmark all your installed models:

```bash
lokai --benchmark
```

Output:
```
⚡ Benchmarking installed models...

   ✓ llama3.1:8b — 42.3 tok/s
   ✓ qwen2.5-coder:7b — 38.7 tok/s
   ✓ deepseek-r1:7b — 35.1 tok/s

📊 Benchmark Results
╭─────────────────────┬──────────┬─────────────┬────────┬────────────╮
│ Model               │ Speed    │ First Token │ Tokens │ Total Time │
├─────────────────────┼──────────┼─────────────┼────────┼────────────┤
│ llama3.1:8b         │ 42.3 t/s │ 0.18s       │ 128    │ 3.2s       │
│ qwen2.5-coder:7b    │ 38.7 t/s │ 0.21s       │ 128    │ 3.5s       │
│ deepseek-r1:7b      │ 35.1 t/s │ 0.24s       │ 128    │ 3.9s       │
╰─────────────────────┴──────────┴─────────────┴────────┴────────────╯

   🏆 Fastest: llama3.1:8b at 42.3 tok/s
```

## Build from Source

```bash
git clone https://github.com/romeo-mz/lokai.git
cd lokai
make build
./bin/lokai
```

### Cross-compile

```bash
make build-all   # Builds for linux/darwin/windows × amd64/arm64
```

### Docker Build

```bash
make docker      # Build Docker image
make docker-run  # Run with Docker
```

## Project Structure

```
cmd/lokai/              → CLI entry point & flags
internal/
  benchmark/            → Model benchmarking engine
  hardware/             → Hardware detection
    ├── detect.go       → Orchestrator (concurrent detection)
    ├── cpu.go          → CPU info (model, cores, AVX)
    ├── memory.go       → RAM detection
    ├── gpu.go          → GPU PCI scan
    ├── gpu_nvidia.go   → NVML VRAM detection
    ├── gpu_amd_linux.go→ sysfs VRAM detection
    └── gpu_apple.go    → Apple Silicon unified memory
  models/
    ├── database.go     → 77-model catalog
    ├── recommend.go    → Recommendation engine
    └── estimate.go     → Performance estimation
  ollama/               → Ollama client wrapper
  ui/                   → Terminal UI (charmbracelet)
```

## Tech Stack

| Component | Library |
|-----------|---------|
| Language | [Go](https://go.dev) |
| Hardware | [ghw](https://github.com/jaypipes/ghw), [gopsutil](https://github.com/shirou/gopsutil), [go-nvml](https://github.com/NVIDIA/go-nvml) |
| LLM Runtime | [Ollama](https://ollama.com/) (native Go API) |
| TUI | [Charmbracelet](https://charm.sh/) (lipgloss, huh) |
| Build | [GoReleaser](https://goreleaser.com/), [Docker](https://www.docker.com/) |
| CI/CD | [GitHub Actions](https://github.com/features/actions) |

## Contributing

Contributions welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

[MIT](LICENSE) — free for personal and commercial use.
