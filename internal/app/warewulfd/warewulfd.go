package warewulfd

import (
	warewulfd_responses "github.com/hpcng/warewulf/internal/app/warewulfd/warewulfd-reponses"
	"github.com/spf13/cobra"
	"net/http"
)

// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp

//const LocalStateDir = "/var/warewulf"


func CobraRunE(cmd *cobra.Command, args []string) error {

	http.HandleFunc("/ipxe/", warewulfd_responses.IpxeSend)
	http.HandleFunc("/kernel/", warewulfd_responses.KernelSend)
	http.HandleFunc("/kmods/", warewulfd_responses.KmodsSend)
	http.HandleFunc("/vnfs/", warewulfd_responses.VnfsSend)
	http.HandleFunc("/overlay-system/", warewulfd_responses.SystemOverlaySend)
	http.HandleFunc("/overlay-runtime", warewulfd_responses.RuntimeOverlaySend)

	http.ListenAndServe(":9873", nil)

	return nil
}
