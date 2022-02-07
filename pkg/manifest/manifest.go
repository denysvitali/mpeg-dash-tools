package manifest

import (
	"fmt"
	"github.com/mc2soft/mpd"
	"net/url"
	"strings"
)

type Manifest struct {
	BaseUrl string
	mpd.MPD
}

func (m *Manifest) GetMedia(t string, lang string) *mpd.AdaptationSet {
	for _, p := range m.MPD.Period {
		for _, a := range p.AdaptationSets {
			if a.ContentType != t {
				continue
			}

			if a.Lang != nil {
				if *a.Lang == lang {
					return a
				}
				continue
			}
			return a
		}
	}
	return nil
}

func (m *Manifest) GetAudio(lang string) *mpd.AdaptationSet {
	return m.GetMedia("audio", lang)
}

func (m *Manifest) GetVideo(lang string) *mpd.AdaptationSet {
	return m.GetMedia("video", lang)
}

func GetBestBandwidth(as *mpd.AdaptationSet) (uint64, mpd.Representation) {
	if as == nil {
		return 0, mpd.Representation{}
	}

	var max uint64 = 0
	var repr mpd.Representation
	for _, v := range as.Representations {
		if v.Bandwidth != nil {
			b := *v.Bandwidth
			if b > max {
				max = b
				repr = v
			}
		}
	}
	return max, repr
}

func (m *Manifest) GetVideoAS(period uint) []*mpd.AdaptationSet {
	if period > uint(len(m.Period))-1 {
		return nil
	}
	p := m.Period[period]
	var videoAS []*mpd.AdaptationSet
	for _, as := range p.AdaptationSets {
		if as.ContentType == "video" {
			videoAS = append(videoAS, as)
		}
	}
	return videoAS
}

func (m *Manifest) GetUrl(
	adapt *mpd.AdaptationSet,
	timestamp uint64,
	bandwidth uint64,
	number uint64,
	repr *mpd.Representation,
) (string, error) {
	if adapt == nil {
		return "", fmt.Errorf("AdaptationSet cannot be nil")
	}

	if adapt.SegmentTemplate == nil {
		return "", fmt.Errorf("AdaptationSet.SegmentTemplate cannot be nil")
	}

	if timestamp == 0 {
		return formatUrl(
			m.BaseUrl,
			adapt.SegmentTemplate.Initialization,
			timestamp,
			number,
			bandwidth,
			repr,
		)
	}
	return formatUrl(
		m.BaseUrl,
		adapt.SegmentTemplate.Media,
		timestamp,
		number,
		bandwidth,
		repr,
	)
}

func formatUrl[T uint64 | int](
	baseUrl string,
	templateUrl *string,
	timestamp uint64,
	mNumber uint64,
	bandwidth T,
	repr *mpd.Representation,
) (string, error) {
	if templateUrl == nil {
		return "", fmt.Errorf("templateUrl cannot be nil")
	}

	u, err := url.Parse(baseUrl)
	if err != nil {
		return "", err
	}

	finalPath := *templateUrl
	finalPath = replaceVariable(finalPath, "Bandwidth", "%d", bandwidth)
	finalPath = replaceVariable(finalPath, "Time", "%d", timestamp)
	finalPath = replaceVariable(finalPath, "RepresentationID", "%s", *repr.ID)
	finalPath = replaceVariable(finalPath, "Number", "%d", mNumber)

	u, err = u.Parse(finalPath)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func replaceVariable[T uint64 | int | string](inputString string, variable string, format string, value T) string {
	return strings.Replace(inputString, "$"+variable+"$", fmt.Sprintf(format, value), -1)
}
