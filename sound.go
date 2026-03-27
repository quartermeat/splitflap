package main

import (
	"math"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	sampleRate = 44100
)

var (
	audioCtx     *audio.Context
	audioCtxOnce sync.Once

	clackSamples   []float32 // base clack PCM — final landing
	flickerSamples []float32 // lighter flicker — intermediate card pass
)

func initAudio() {
	audioCtxOnce.Do(func() {
		audioCtx = audio.NewContext(sampleRate)
		clackSamples = generateClack()
		flickerSamples = generateFlicker()
	})
}

// generateClack synthesizes a single split-flap clack sound.
//
// Real split-flap clack anatomy:
//  1. Sharp transient — plastic card hitting the stop pin (~2ms)
//  2. Resonant body — the card resonating briefly (~15ms)
//  3. Fast noise tail — air turbulence from the flip (~10ms)
func generateClack() []float32 {
	durationMs := 35
	n := sampleRate * durationMs / 1000
	samples := make([]float32, n)

	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate

		// 1. Sharp transient click (first 2ms) — damped tone at ~800Hz.
		transient := 0.0
		if t < 0.002 {
			decay := math.Exp(-t * 800)
			transient = decay * math.Sin(2*math.Pi*800*t)
		}

		// 2. Card resonance (2–18ms) — 200Hz body thump, decays fast.
		resonance := 0.0
		if t >= 0.002 && t < 0.018 {
			rt := t - 0.002
			decay := math.Exp(-rt * 300)
			resonance = decay * math.Sin(2*math.Pi*200*rt) * 0.6
		}

		// 3. White noise tail (0–25ms) — the flutter/turbulence.
		noise := 0.0
		if t < 0.025 {
			noiseDecay := math.Exp(-t * 200)
			noise = noiseDecay * (rand.Float64()*2 - 1) * 0.25
		}

		samples[i] = float32(transient + resonance + noise)
	}

	// Normalize.
	peak := float32(0)
	for _, s := range samples {
		if s < 0 {
			s = -s
		}
		if s > peak {
			peak = s
		}
	}
	if peak > 0 {
		scale := float32(0.85) / peak
		for i := range samples {
			samples[i] *= scale
		}
	}

	return samples
}

// pitchShift returns a copy of samples with a simple pitch shift by resampling.
func pitchShift(samples []float32, factor float64) []float32 {
	newLen := int(float64(len(samples)) / factor)
	out := make([]float32, newLen)
	for i := range out {
		src := float64(i) * factor
		idx := int(src)
		frac := float32(src - float64(idx))
		if idx+1 < len(samples) {
			out[i] = samples[idx]*(1-frac) + samples[idx+1]*frac
		} else if idx < len(samples) {
			out[i] = samples[idx]
		}
	}
	return out
}

// generateFlicker synthesizes the brief sound of a card passing through mid-cycle.
// Much shorter and softer than a clack — just the transient + a tiny noise burst.
func generateFlicker() []float32 {
	durationMs := 12
	n := sampleRate * durationMs / 1000
	samples := make([]float32, n)

	for i := 0; i < n; i++ {
		t := float64(i) / sampleRate

		// Quick transient tick — higher frequency, faster decay than full clack.
		transient := 0.0
		if t < 0.001 {
			decay := math.Exp(-t * 2000)
			transient = decay * math.Sin(2*math.Pi*1200*t) * 0.5
		}

		// Very short noise burst.
		noise := 0.0
		if t < 0.008 {
			decay := math.Exp(-t * 600)
			noise = decay * (rand.Float64()*2 - 1) * 0.15
		}

		samples[i] = float32(transient + noise)
	}

	// Normalize to ~40% of full volume — noticeably quieter than the landing clack.
	peak := float32(0)
	for _, s := range samples {
		if s < 0 {
			s = -s
		}
		if s > peak {
			peak = s
		}
	}
	if peak > 0 {
		scale := float32(0.40) / peak
		for i := range samples {
			samples[i] *= scale
		}
	}

	return samples
}

func playSamples(samples []float32) {
	if audioCtx == nil {
		return
	}
	factor := 0.92 + rand.Float64()*0.16
	shifted := pitchShift(samples, factor)

	buf := make([]byte, len(shifted)*4)
	for i, s := range shifted {
		v := int16(s * 32767)
		buf[i*4] = byte(v)
		buf[i*4+1] = byte(v >> 8)
		buf[i*4+2] = byte(v)
		buf[i*4+3] = byte(v >> 8)
	}
	audioCtx.NewPlayerFromBytes(buf).Play()
}

// PlayClack plays the final landing clack (hard stop).
func PlayClack() {
	playSamples(clackSamples)
}

// PlayFlicker plays the light tick of an intermediate card passing through.
func PlayFlicker() {
	playSamples(flickerSamples)
}
