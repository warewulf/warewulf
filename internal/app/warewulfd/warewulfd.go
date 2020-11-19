package warewulfd

import (
	"github.com/hpcng/warewulf/internal/app/warewulfd/response"
	"github.com/spf13/cobra"
	"net/http"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp

func CobraRunE(cmd *cobra.Command, args []string) error {

	http.HandleFunc("/ipxe/", response.IpxeSend)
	http.HandleFunc("/kernel/", response.KernelSend)
	http.HandleFunc("/kmods/", response.KmodsSend)
	http.HandleFunc("/vnfs/", response.VnfsSend)
	http.HandleFunc("/overlay-system/", response.SystemOverlaySend)
	http.HandleFunc("/overlay-runtime", response.RuntimeOverlaySend)

	http.ListenAndServe(":9873", nil)

	return nil
}
