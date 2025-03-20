package utils

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/h2non/filetype"
)

const (
	// dlnaOrgFlagSenderPaced = 1 << 31
	// dlnaOrgFlagTimeBasedSeek = 1 << 30
	// dlnaOrgFlagByteBasedSeek = 1 << 29
	// dlnaOrgFlagPlayContainer = 1 << 28
	// dlnaOrgFlagS0Increase = 1 << 27
	// dlnaOrgFlagSnIncrease = 1 << 26
	// dlnaOrgFlagRtspPause = 1 << 25
	dlnaOrgFlagStreamingTransferMode = 1 << 24
	// dlnaOrgFlagInteractiveTransfertMode = 1 << 23
	dlnaOrgFlagBackgroundTransfertMode = 1 << 22
	dlnaOrgFlagConnectionStall         = 1 << 21
	dlnaOrgFlagDlnaV15                 = 1 << 20
)

var (
	dlnaprofiles = map[string]string{
		"video/x-mkv":             "DLNA.ORG_PN=MATROSKA",
		"video/x-matroska":        "DLNA.ORG_PN=MATROSKA",
		"video/x-msvideo":         "DLNA.ORG_PN=AVI",
		"video/mpeg":              "DLNA.ORG_PN=MPEG1",
		"video/vnd.dlna.mpeg-tts": "DLNA.ORG_PN=MPEG1",
		"video/mp4":               "DLNA.ORG_PN=AVC_MP4_MP_SD_AAC_MULT5",
		"video/quicktime":         "DLNA.ORG_PN=AVC_MP4_MP_SD_AAC_MULT5",
		"video/x-m4v":             "DLNA.ORG_PN=AVC_MP4_MP_SD_AAC_MULT5",
		"video/3gpp":              "DLNA.ORG_PN=AVC_MP4_MP_SD_AAC_MULT5",
		"video/x-flv":             "DLNA.ORG_PN=AVC_MP4_MP_SD_AAC_MULT5",
		"video/x-ms-wmv":          "DLNA.ORG_PN=WMVHIGH_FULL",
		"audio/mpeg":              "DLNA.ORG_PN=MP3",
		"image/jpeg":              "JPEG_LRG",
		"image/png":               "PNG_LRG",
	}

	ErrInvalidSeekFlag    = errors.New("invalid seek flag")
	ErrInvalidClockFormat = errors.New("invalid clock format")
)

func defaultStreamingFlags() string {
	return fmt.Sprintf("%.8x%.24x", dlnaOrgFlagStreamingTransferMode|
		dlnaOrgFlagBackgroundTransfertMode|
		dlnaOrgFlagConnectionStall|
		dlnaOrgFlagDlnaV15, 0)
}

func BuildRequestHeader(soapAction string) http.Header {
	return http.Header{
		"SOAPAction":   []string{soapAction},
		"content-type": []string{"text/xml"},
		"charset":      []string{"utf-8"},
		"Connection":   []string{"close"},
	}
}

// BuildContentFeatures builds the content features string
// for the "contentFeatures.dlna.org" header.
func BuildContentFeatures(mediaType string, seek string, transcode bool) (string, error) {
	var cf strings.Builder

	if mediaType != "" {
		dlnaProf, profExists := dlnaprofiles[mediaType]
		if profExists {
			cf.WriteString(dlnaProf + ";")
		}
	}

	// "00" neither time seek range nor range supported
	// "01" range supported
	// "10" time seek range supported
	// "11" both time seek range and range supported
	switch seek {
	case "00":
		cf.WriteString("DLNA.ORG_OP=00;")
	case "01":
		cf.WriteString("DLNA.ORG_OP=01;")
	case "10":
		cf.WriteString("DLNA.ORG_OP=10;")
	case "11":
		cf.WriteString("DLNA.ORG_OP=11;")
	default:
		return "", ErrInvalidSeekFlag
	}

	switch transcode {
	case true:
		cf.WriteString("DLNA.ORG_CI=1;")
	default:
		cf.WriteString("DLNA.ORG_CI=0;")
	}

	cf.WriteString("DLNA.ORG_FLAGS=")
	cf.WriteString(defaultStreamingFlags())

	return cf.String(), nil
}

// GetMimeDetails returns the media mime details.
func GetMimeDetails(f io.ReadCloser) (string, error) {
	defer f.Close()
	head := make([]byte, 261)
	_, err := f.Read(head)
	if err != nil {
		return "", fmt.Errorf("getMimeDetailsFromFile error #2: %w", err)
	}

	kind, err := filetype.Match(head)
	if err != nil {
		return "", fmt.Errorf("getMimeDetailsFromFile error #3: %w", err)
	}

	return fmt.Sprintf("%s/%s", kind.MIME.Type, kind.MIME.Subtype), nil
}

// ClockTimeToSeconds converts relative time to seconds.
func ClockTimeToSeconds(strtime string) (int, error) {
	var out int
	v := make([]int, 0, 3)

	s := strings.Split(strtime, ":")
	if len(s) != 3 {
		return 0, ErrInvalidClockFormat
	}

	num, err := strconv.Atoi(s[0])
	if err != nil {
		return 0, ErrInvalidClockFormat
	}
	v = append(v, num)

	num, err = strconv.Atoi(s[1])
	if err != nil {
		return 0, ErrInvalidClockFormat
	}
	v = append(v, num)

	f, err := strconv.ParseFloat(s[2], 32)
	if err != nil {
		return 0, ErrInvalidClockFormat
	}
	f = math.Round(f)
	v = append(v, int(f))

	for n, i := range v {
		switch n {
		case 0:
			out += i * 3600
		case 1:
			out += i * 60
		case 2:
			out += i
		}
	}

	return out, nil
}

// SecondsToClockTime converts seconds to seconds relative time.
func SecondsToClockTime(secs int) (string, error) {
	hours := secs / 3600
	secs %= 3600
	minutes := secs / 60
	secs %= 60

	str := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
	return str, nil
}

// FormatClockTime converts clock time to a more expected format of clock time.
func FormatClockTime(strtime string) (string, error) {
	sec, err := ClockTimeToSeconds(strtime)
	if err != nil {
		return "", ErrInvalidClockFormat
	}

	out, err := SecondsToClockTime(sec)
	if err != nil {
		return "", ErrInvalidClockFormat
	}

	return out, nil
}

// ParseDuration parses a HH:MM:SS or MM:SS formatted string
// into a [time.Duration] value.
func ParseDuration(durStr string) (time.Duration, error) {
	parts := strings.Split(durStr, ":")
	var hours, minutes, seconds int
	var err error

	if len(parts) == 2 {
		// Format MM:SS
		minutes, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		seconds, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %w", err)
		}
	} else if len(parts) == 3 {
		// Format HH:MM:SS
		hours, err = strconv.Atoi(parts[0])
		if err != nil {
			return 0, fmt.Errorf("invalid hours: %w", err)
		}
		minutes, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, fmt.Errorf("invalid minutes: %w", err)
		}
		seconds, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, fmt.Errorf("invalid seconds: %w", err)
		}
	} else {
		return 0, fmt.Errorf("invalid format: expected MM:SS or HH:MM:SS")
	}

	duration := time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second
	return duration, nil
}
