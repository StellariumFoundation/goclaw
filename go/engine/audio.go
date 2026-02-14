// audio.go — Sound and music playback.
//
// Wraps Ebitengine's audio API to provide simple play/stop/loop controls
// for sound effects and background music. Captain Claw uses WAV and MIDI
// formats; this layer will handle decoding and playback.
package engine

import "github.com/hajimehoshi/ebiten/v2/audio"

const sampleRate = 44100

// AudioManager manages the audio context and active players.
type AudioManager struct {
	context *audio.Context
	// TODO: track active sound effect and music players
}

// NewAudioManager initializes the audio subsystem.
func NewAudioManager() *AudioManager {
	ctx := audio.NewContext(sampleRate)
	return &AudioManager{
		context: ctx,
	}
}

// PlaySFX plays a one-shot sound effect.
// TODO: implement loading from decoded audio bytes.
func (am *AudioManager) PlaySFX(name string) {
	// placeholder — will load and play sound effects from asset cache
}

// PlayMusic starts looping background music.
// TODO: implement music streaming.
func (am *AudioManager) PlayMusic(name string) {
	// placeholder — will stream and loop background music tracks
}

// StopMusic stops the currently playing background music.
func (am *AudioManager) StopMusic() {
	// placeholder
}
