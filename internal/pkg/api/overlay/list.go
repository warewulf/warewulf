package apioverlay

import (
	"os"
	"strconv"
	"strings"
	"syscall"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/overlay"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/pkg/errors"
)

func OverlayList(overlayGet *wwapiv1.OverlayListParameter) (*overlay.OverlayListResponse, error) {
	var overlays []string

	if len(overlayGet.Overlays) > 0 {
		overlays = overlayGet.Overlays
	} else {
		var err error
		overlays, err = overlay.FindOverlays()
		if err != nil {
			return nil, errors.Wrap(err, "could not obtain list of overlays from system")
		}
	}

	resp := &overlay.OverlayListResponse{
		Overlays: make(map[string][]overlay.OverlayListEntry),
	}

	if strings.EqualFold(overlayGet.Type.String(), "TYPE_LONG") {
		// this is long format
		for o := range overlays {
			name := overlays[o]
			path := overlay.OverlaySourceDir(name)

			if util.IsDir(path) {
				files := util.FindFiles(path)

				var entries []overlay.OverlayListEntry
				for file := range files {
					s, err := os.Stat(files[file])
					if err != nil {
						continue
					}

					fileMode := s.Mode()
					perms := fileMode & os.ModePerm

					sys := s.Sys()

					entries = append(entries, &overlay.OverlayListLongEntry{
						PermMode: perms.String(),
						UID:      strconv.FormatUint(uint64(sys.(*syscall.Stat_t).Uid), 10),
						GID:      strconv.FormatUint(uint64(sys.(*syscall.Stat_t).Gid), 10),
						FilePath: files[file],
					})
				}

				if vals, ok := resp.Overlays[overlays[o]]; ok {
					entries = append(entries, vals...)
				}
				resp.Overlays[overlays[o]] = entries
			}
		}
	} else {
		for o := range overlays {
			name := overlays[o]
			path := overlay.OverlaySourceDir(name)

			if util.IsDir(path) {
				files := util.FindFiles(path)

				var entries []overlay.OverlayListEntry
				if strings.EqualFold(overlayGet.Type.String(), "TYPE_CONTENT") {
					// it'll be TYPE_TYPE_CONTENT
					if len(files) == 0 {
						entries = append(entries, &overlay.OverlayListSimpleEntry{
							FilesDirs: "0",
						})
					} else {
						for file := range files {
							entries = append(entries, &overlay.OverlayListSimpleEntry{
								FilesDirs: files[file],
							})
						}
					}
				} else {
					// for unspecific version, it'll be default format
					entries = append(entries, &overlay.OverlayListSimpleEntry{
						FilesDirs: strconv.Itoa(len(files)),
					})
				}

				if vals, ok := resp.Overlays[name]; ok {
					entries = append(entries, vals...)
				}
				resp.Overlays[name] = entries
			}
		}
	}

	return resp, nil
}
