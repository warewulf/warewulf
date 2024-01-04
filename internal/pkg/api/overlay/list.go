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

func OverlayList(overlayGet *wwapiv1.OverlayListParameter) (*wwapiv1.OverlayListResponse, error) {
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

	var entries []*wwapiv1.OverlayListResponseEntry
	if strings.EqualFold(overlayGet.Type.String(), "TYPE_LONG") {
		// this is long format
		for o := range overlays {
			name := overlays[o]
			path := overlay.OverlaySourceDir(name)

			if util.IsDir(path) {
				files := util.FindFiles(path)

				for file := range files {
					s, err := os.Stat(files[file])
					if err != nil {
						continue
					}

					fileMode := s.Mode()
					perms := fileMode & os.ModePerm

					sys := s.Sys()

					entries = append(entries, &wwapiv1.OverlayListResponseEntry{
						OverlayEntry: &wwapiv1.OverlayListResponseEntry_OverlayLong{
							OverlayLong: &wwapiv1.OverlayListLong{
								PermMode:      perms.String(),
								Uid:           strconv.FormatUint(uint64(sys.(*syscall.Stat_t).Uid), 10),
								Gid:           strconv.FormatUint(uint64(sys.(*syscall.Stat_t).Gid), 10),
								SystemOverlay: overlays[o],
								FilePath:      files[file],
							},
						},
					})
				}
			}
		}
	} else {
		for o := range overlays {
			name := overlays[o]
			path := overlay.OverlaySourceDir(name)

			if util.IsDir(path) {
				files := util.FindFiles(path)

				if strings.EqualFold(overlayGet.Type.String(), "TYPE_TYPE_UNSPECIFIED") {
					// for unspecific version, it'll be default format
					entries = append(entries, &wwapiv1.OverlayListResponseEntry{
						OverlayEntry: &wwapiv1.OverlayListResponseEntry_OverlaySimple{
							OverlaySimple: &wwapiv1.OverlayListSimple{
								OverlayName: name,
								FilesDirs:   strconv.Itoa(len(files)),
							},
						},
					})
				} else {
					// otherwise, it'll be TYPE_TYPE_CONTENT
					if len(files) == 0 {
						entries = append(entries, &wwapiv1.OverlayListResponseEntry{
							OverlayEntry: &wwapiv1.OverlayListResponseEntry_OverlaySimple{
								OverlaySimple: &wwapiv1.OverlayListSimple{
									OverlayName: name,
									FilesDirs:   "0",
								},
							},
						})
					} else {
						for file := range files {
							entries = append(entries, &wwapiv1.OverlayListResponseEntry{
								OverlayEntry: &wwapiv1.OverlayListResponseEntry_OverlaySimple{
									OverlaySimple: &wwapiv1.OverlayListSimple{
										OverlayName: name,
										FilesDirs:   files[file],
									},
								},
							})
						}
					}
				}
			}
		}
	}

	return &wwapiv1.OverlayListResponse{Overlays: entries}, nil
}
