package mpeg_dash_tools

import (
	"encoding/xml"
	"fmt"
	"github.com/denysvitali/mpeg-dash-tools/pkg/manifest"
	"github.com/mc2soft/mpd"
	"github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
	"time"
)

type DumpClient struct {
	client HttpClient
	logger *logrus.Logger
}

func NewDumpClient() *DumpClient {
	crt := CustomRoundTripper{embeddedRoundTripper: http.DefaultTransport}
	c := http.Client{
		Timeout:   10 * time.Second,
		Transport: crt,
	}
	d := DumpClient{client: &c, logger: logrus.New()}
	return &d
}

func (d *DumpClient) SetLogger(l *logrus.Logger) {
	if l != nil {
		d.logger = l
	}
}

func getManifestPath(s string) string {
	return fmt.Sprintf("%s", s)
}

func (d *DumpClient) FetchManifest(u string) (*manifest.Manifest, error) {
	req, err := http.NewRequest(http.MethodGet, getManifestPath(u), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %v", err)
	}

	res, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"invalid status code %s: 200 OK expected",
			res.Status,
		)
	}

	var mpdManifest mpd.MPD
	dec := xml.NewDecoder(res.Body)
	err = dec.Decode(&mpdManifest)

	m := manifest.Manifest{
		BaseUrl: u,
	}
	m.MPD = mpdManifest
	return &m, err
}

func (d *DumpClient) Process(u string) {
	m, err := d.FetchManifest(u)
	if err != nil {
		d.logger.Fatalf("unable to fetch manifest: %v", err)
	}

	videoAs := m.GetVideoAS(0)

	// Find best video
	var maxSurface uint64 = 0
	var bestVideo = videoAs[0]
	for _, v := range videoAs {
		if v.MaxWidth == nil || v.MaxHeight == nil {
			continue
		}
		surfaceSize := *v.MaxWidth * *v.MaxHeight
		if surfaceSize > maxSurface {
			maxSurface = surfaceSize
			bestVideo = v
		}
	}

	if bestVideo.MaxWidth != nil && bestVideo.MaxHeight != nil {
		d.logger.Printf("best video (%d x %d)", bestVideo.MaxWidth, bestVideo.MaxHeight)
	}

	tmpDir, err := os.MkdirTemp(os.TempDir(), "video-mpd-*")
	if err != nil {
		d.logger.Fatalf("unable to create temporary dir: %v", err)
	}

	var repr mpd.Representation
	var bestSurface = uint64(0)
	for _, v := range bestVideo.Representations {
		surface := *v.Width * *v.Height
		if surface > bestSurface {
			bestSurface = surface
			repr = v
		}
	}

	// Get init frame
	// Get best bandwidth
	bestBandwidth, _ := manifest.GetBestBandwidth(bestVideo)
	d.saveFrameAtTimestamp(0, 0, tmpDir, m, bestVideo, bestBandwidth, &repr)
	var startTs uint64 = 0
	var timeScale uint64 = 1000
	var duration uint64
	if bestVideo.SegmentTemplate != nil {
		if bestVideo.SegmentTemplate.PresentationTimeOffset != nil {
			startTs += *bestVideo.SegmentTemplate.PresentationTimeOffset
		}
		if bestVideo.SegmentTemplate.Timescale != nil {
			timeScale = *bestVideo.SegmentTemplate.Timescale
		}

		if bestVideo.SegmentTemplate.Duration != nil {
			duration = *bestVideo.SegmentTemplate.Duration
		}
	}

	offset := startTs
	if bestVideo.SegmentTemplate.SegmentTimeline != nil {
		segmentTimeline := bestVideo.SegmentTemplate.SegmentTimeline.S[0]
		d.logger.Debugf("timeScale=%d, startTs=%d, offset=%d", timeScale, startTs, offset)

		repetitions := uint64(*segmentTimeline.R)
		duration := segmentTimeline.D
		for i := uint64(0); i < repetitions; i++ {
			d.saveFrameAtTimestamp(offset+i*duration, i, tmpDir, m, bestVideo, bestBandwidth, &repr)
		}
	} else {
		for i := uint64(1); i < duration; i++ {
			d.saveFrameAtTimestamp(offset+i*duration*timeScale, i, tmpDir, m, bestVideo, bestBandwidth, &repr)
		}
	}
}

func (d *DumpClient) saveFrameAtTimestamp(
	ts uint64,
	number uint64,
	tmpDir string,
	m *manifest.Manifest,
	bestVideo *mpd.AdaptationSet,
	bestBandwidth uint64,
	repr *mpd.Representation,
) {
	u, err := m.GetUrl(bestVideo, ts, bestBandwidth, number, repr)
	if err != nil {
		d.logger.Fatalf("unable to get url: %v", err)
	}

	d.logger.Debugf("url is %v", u)

	err = d.saveFrame(u, ts, tmpDir)
	if err != nil {
		d.logger.Fatalf("unable to save frame: %v", err)
	}
}

func (d *DumpClient) saveFrame(u string, idx uint64, tmpDir string) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}

	res, err := d.client.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %s: 200 OK expected", res.Status)
	}

	//fileName := path.Join(tmpDir, fmt.Sprintf("%d.bin", idx))
	fileName := path.Join(tmpDir, "output.bin")
	f, err := os.OpenFile(
		fileName,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	d.logger.Debugf("creating file %s", fileName)
	defer f.Close()
	if err != nil {
		return err
	}

	// Content is base64 encoded, let's decode it
	// dec := base64.NewDecoder(base64.StdEncoding, res.Body)
	_, err = io.Copy(f, res.Body)
	if err != nil {
		return err
	}
	return nil
}
