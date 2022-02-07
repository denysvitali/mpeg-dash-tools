package mpeg_dash_tools

import (
	"fmt"
	"github.com/denysvitali/mpeg-dash-tools/pkg/manifest"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"net/http"
	"os"
	"testing"
)

func TestFetchManifest(t *testing.T) {
	err, mpd := getMockedManifest(t, "./test/mpd/1.xml")
	assert.Nil(t, err)
	assert.NotNil(t, mpd)
}

func getMockedManifest(t *testing.T, filePath string) (*manifest.Manifest, error) {
	var responses []http.Response
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("unable to open mocked response: %v", err)
	}
	responses = append(responses, http.Response{
		StatusCode: http.StatusOK,
		Body:       f,
	})
	d := DumpClient{client: &MockedHttpClient{responses: responses}}
	return d.FetchManifest("https://example.com/dash/KOMJILAAIKEMPPBG.mpd")
}

func TestGetFragmentList(t *testing.T) {
	mpd, err := getMockedManifest(t, "./test/mpd/1.xml")
	if err != nil {
		t.Fatalf("unable to get mocked manifest: %v", err)
	}

	for i, p := range mpd.Period {
		fmt.Printf("Period %d:\n", i)
		fmt.Printf("\tDuration: %d:\n", p.Duration)
		fmt.Printf("\tAdaptation Sets:\n")
		for i2, as := range p.AdaptationSets {
			fmt.Printf("\t\tAdaptation Set %d\n", i2)
			printIfNotNil(3, "Codecs", as.Codecs)
			printIfNotNil(3, "Lang", as.Lang)
			printIfNotNil(3, "MimeType", &as.MimeType)

			if as.SegmentTemplate != nil {
				st := as.SegmentTemplate
				printIfNotNil(3, "ST Media", st.Media)
				printIfNotNil(3, "ST Init", st.Initialization)

				if st.SegmentTimeline != nil {
					var segmentTime uint64

					for _, v := range st.SegmentTimeline.S {
						if v.T != nil {
							segmentTime += *v.T
						}
						segmentEnd := segmentTime + v.D
						fmt.Printf(
							"\t\t\tST Segment End: %d\n",
							segmentEnd,
						)
						segmentTime = segmentEnd
					}
				}
				fmt.Printf("\t\t\tST StartNumber %d\n", st.StartNumber)
				fmt.Printf("\t\t\tST Timescale %d\n", st.Timescale)
			}

			for i3, repr := range as.Representations {
				fmt.Printf("\t\t\tRepresentation %d: \n", i3)
				printIfNotNil(4, "ID", repr.ID)
				printIfNotNil(4, "Codecs", repr.Codecs)
				printIfNotNil(4, "FrameRate", repr.FrameRate)
				printIfNotNil(4, "BaseURL", repr.BaseURL)
				printIfNotNil(4, "Bandwidth", repr.Bandwidth)
			}
		}
	}
}

func printIfNotNil[T string | uint64](indent int, key string, value *T) {
	if value == nil {
		return
	}
	for i := 0; i < indent; i++ {
		fmt.Printf("\t")
	}
	fmt.Printf("%s: %v\n", key, *value)
}

func fileResponse(t *testing.T, filePath string) http.Response {
	f, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("unable to open file: %v", err)
		return http.Response{}
	}

	return http.Response{StatusCode: http.StatusOK, Body: f}
}

func TestDumpClient_Process(t *testing.T) {
	m := MockedHttpClient{
		requests: []RequestTuple{
			{Method: http.MethodGet, Url: "https://example.com/dash/file.mpd/Manifest"},
			{Method: http.MethodGet, Url: "https://example.com/dash/file.mpd/QualityLevels(3200000)/Fragments(video=Init)"},
			{Method: http.MethodGet, Url: "https://example.com/dash/file.mpd/QualityLevels(3200000)/Fragments(video=16440052800000000)"},
			{Method: http.MethodGet, Url: "https://example.com/dash/file.mpd/QualityLevels(3200000)/Fragments(video=16440052840000000)"},
			{Method: http.MethodGet, Url: "https://example.com/dash/file.mpd/QualityLevels(3200000)/Fragments(video=16440052880000000)"},
			{Method: http.MethodGet, Url: "https://example.com/dash/file.mpd/QualityLevels(3200000)/Fragments(video=16440052920000000)"},
		},
		responses: []http.Response{
			fileResponse(t, "./test/mpd/1.xml"),
			fileResponse(t, "./test/mpd/fragments/init.txt"),
			fileResponse(t, "./test/mpd/fragments/1.txt"),
			fileResponse(t, "./test/mpd/fragments/1.txt"),
			fileResponse(t, "./test/mpd/fragments/1.txt"),
			fileResponse(t, "./test/mpd/fragments/1.txt"),
		},
	}
	d := DumpClient{client: &m, logger: logrus.New()}
	d.logger.SetLevel(logrus.DebugLevel)
	d.Process("https://example.com/dash/file.mpd")
}
