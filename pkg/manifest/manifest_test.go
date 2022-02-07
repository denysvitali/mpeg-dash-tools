package manifest_test

import (
	"encoding/xml"
	manifest "github.com/denysvitali/mpeg-dash-tools/pkg/manifest"
	"github.com/mc2soft/mpd"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestManifest_GetAudio(t *testing.T) {
	m := loadSampleManifest(t, "./test/mpd/1.xml")
	audio := m.GetAudio("ita")
	assert.NotNil(t, audio)
}

func TestManifest_GetUrl(t *testing.T) {
	m := loadSampleManifest(t, "./test/mpd/1.xml")
	audio := m.GetAudio("ita")
	u, err := m.GetUrl(audio, 0, *audio.Representations[0].Bandwidth)
	assert.Nil(t, err)
	assert.Equal(t, "https://example.com/dash/example.mpd/QualityLevels(96000)/Fragments(audio_482_ita=Init)", u)

	u, err = m.GetUrl(audio, 39456133248533, *audio.Representations[0].Bandwidth)
	assert.Equal(t, "https://example.com/dash/example.mpd/QualityLevels(96000)/Fragments(audio_482_ita=39456133248533)", u)

	video := m.GetVideo("ita")
	assert.NotNil(t, video)
	bb, _ := manifest.GetBestBandwidth(video)
	u, err = m.GetUrl(video, 0, bb)
	assert.Nil(t, err)
	assert.Equal(t, "https://example.com/dash/example.mpd/QualityLevels(3200000)/Fragments(video=Init)", u)
}

func loadSampleManifest(t *testing.T, manifestFile string) manifest.Manifest {
	f, err := os.Open(manifestFile)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
	}

	var mpd mpd.MPD
	dec := xml.NewDecoder(f)
	err = dec.Decode(&mpd)
	if err != nil {
		t.Fatalf("unable to decode XML: %v", err)
	}

	return manifest.Manifest{
		MPD:     mpd,
		BaseUrl: "https://example.com/dash/example.mpd",
	}
}
