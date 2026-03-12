package warewulfd

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// buildTemplateVars constructs the templateVars struct with all necessary
// fields, including handling IPv6 authority formatting and kernel version resolution.
func buildTemplateVars(conf *warewulfconf.WarewulfYaml, rinfo parsedRequest, remoteNode node.Node) *templateVars {
	kernelArgs := ""
	kernelVersion := ""
	if remoteNode.Kernel != nil {
		kernelArgs = strings.Join(remoteNode.Kernel.Args, " ")
		kernelVersion = remoteNode.Kernel.Version
	}
	if kernelVersion == "" {
		if kernel_ := kernel.FromNode(&remoteNode); kernel_ != nil {
			kernelVersion = kernel_.Version()
		}
	}

	authority := fmt.Sprintf("%s:%d", conf.Ipaddr, conf.Warewulf.Port)
	ipaddr6 := ""
	if confIpaddr6, err := netip.ParseAddr(conf.Ipaddr6); err == nil {
		ipaddr6 = confIpaddr6.String()
	}
	if rinfoIpaddr, err := netip.ParseAddr(rinfo.ipaddr); err == nil {
		if rinfoIpaddr.Is6() {
			if ipaddr6 != "" {
				authority = fmt.Sprintf("[%s]:%d", ipaddr6, conf.Warewulf.Port)
			} else {
				wwlog.Error("No valid IPv6 address configured, but request is IPv6")
			}
		}
	} else {
		wwlog.Error("Could not parse request IP address: %s", rinfo.ipaddr)
	}

	return &templateVars{
		Id:            remoteNode.Id(),
		Cluster:       remoteNode.ClusterName,
		Fqdn:          remoteNode.Id(),
		Ipaddr:        conf.Ipaddr,
		Ipaddr6:       ipaddr6,
		Port:          strconv.Itoa(conf.Warewulf.Port),
		TLS:           conf.Warewulf.TLSEnabled(),
		Authority:     authority,
		Hostname:      remoteNode.Id(),
		Hwaddr:        rinfo.hwaddr,
		ImageName:     remoteNode.ImageName,
		Ipxe:          remoteNode.Ipxe,
		KernelArgs:    kernelArgs,
		KernelVersion: kernelVersion,
		Root:          remoteNode.Root,
		NetDevs:       remoteNode.NetDevs,
		Tags:          remoteNode.Tags}
}

// sendResponse handles the common response logic for provision handlers.
// If tmplData is non-nil, it renders the stageFile as a template. Otherwise, it
// sends stageFile as a raw file (with optional .gz compression).
func sendResponse(w http.ResponseWriter, req *http.Request, stageFile string, tmplData *templateVars, ctx *requestContext) {
	wwlog.Serv("stage_file '%s'", stageFile)

	if util.IsFile(stageFile) {

		if tmplData != nil {
			if ctx.rinfo.compress != "" {
				wwlog.Error("Unsupported %s compressed version for file: %s",
					ctx.rinfo.compress, stageFile)
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Create a template with the Sprig functions.
			tmpl := template.New(filepath.Base(stageFile)).Funcs(sprig.TxtFuncMap())

			// Parse the template.
			parsedTmpl, err := tmpl.ParseFiles(stageFile)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				wwlog.ErrorExc(err, "")
				return
			}

			// template engine writes file to buffer in case rendering fails
			var buf bytes.Buffer

			err = parsedTmpl.Execute(&buf, tmplData)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				wwlog.ErrorExc(err, "")
				return
			}

			w.Header().Set("Content-Type", "text")
			w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
			ctx.tpm.Update(stageFile, fmt.Sprintf("%x", sha256.Sum256(buf.Bytes())))

			_, err = buf.WriteTo(w)
			if err != nil {
				wwlog.ErrorExc(err, "")
			}

			wwlog.Info("send %s -> %s", stageFile, ctx.remoteNode.Id())

		} else {
			if ctx.rinfo.compress == "gz" {
				stageFile += ".gz"

				if !util.IsFile(stageFile) {
					wwlog.Error("unprepared for compressed version of file %s",
						stageFile)
					w.WriteHeader(http.StatusNotFound)
					return
				}
			} else if ctx.rinfo.compress != "" {
				wwlog.Error("unsupported %s compressed version of file %s",
					ctx.rinfo.compress, stageFile)
				w.WriteHeader(http.StatusNotFound)
			}
			// Read file content for checksum
			fileBytes, err := os.ReadFile(stageFile)
			if err != nil {
				wwlog.ErrorExc(err, "")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			ctx.tpm.Update(stageFile, fmt.Sprintf("%x", sha256.Sum256(fileBytes)))

			err = sendFile(w, req, stageFile, ctx.remoteNode.Id())
			if err != nil {
				wwlog.ErrorExc(err, "")
				return
			}
		}

		updateStatus(ctx.remoteNode.Id(), ctx.rinfo.stage, path.Base(stageFile), ctx.rinfo.ipaddr)

	} else if stageFile == "" {
		w.WriteHeader(http.StatusBadRequest)
		wwlog.Error("No resource selected")
		updateStatus(ctx.remoteNode.Id(), ctx.rinfo.stage, "BAD_REQUEST", ctx.rinfo.ipaddr)

	} else {
		w.WriteHeader(http.StatusNotFound)
		wwlog.Error("Not found: %s", stageFile)
		updateStatus(ctx.remoteNode.Id(), ctx.rinfo.stage, "NOT_FOUND", ctx.rinfo.ipaddr)
	}
}
