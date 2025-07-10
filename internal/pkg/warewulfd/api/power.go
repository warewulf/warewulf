package api

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/bmc"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

// PowerResult is the result structure for single node power status operations. get/set power status
type PowerResult struct {
	NodeId string `yaml:"node_id" json:"node_id" description:"ID of the node for which power status is being reported"`
	BmcIp  string `yaml:"bmc_ip" json:"bmc_ip" description:"BMC IP address of the node"`
	Result string `yaml:"result" json:"result" description:"Power status of the node, e.g., 'on', 'off', 'unknown'"`
	Err    error  `yaml:"error,omitempty" json:"error,omitempty" description:"Error message if any error occurred while getting power status"`
}

// getPower returns an interactor to get the power status of a node by its ID.
// At this time we ony support a single node. There is no expansion of the hostlist args or batching yet.
func getPower() usecase.Interactor {
	type getPowerInput struct {
		ID string `path:"id" required:"true" description:"ID of node to get power status for"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getPowerInput, output *[]PowerResult) error {
		wwlog.Debug("api.getPower(), ID: %s", input.ID)
		powerResults := []PowerResult{}
		powerResults = append(powerResults, PowerResult{ // start with one expansion is not yet supported
			NodeId: input.ID,
		})

		if registry, err := node.New(); err != nil {
			wwlog.Debug("api.getPower() getting node registry: %v", err)
			powerResults[0].Err = err // Set the error in the first PowerResult
			*output = powerResults    // Set *output on exit. important
			return err
		} else {
			if node_, err := registry.GetNode(input.ID); err != nil {
				wwlog.Debug("api.getPower() node not found: %v", err)
				powerResults[0].Err = err
				*output = powerResults // Set *output on exit. important
				return status.Wrap(fmt.Errorf("node not found: %v (%v)", input.ID, err), status.NotFound)
			} else {
				if node_.Ipmi == nil || node_.Ipmi.Ipaddr.IsUnspecified() || node_.Ipmi.UserName == "" || node_.Ipmi.Password == "" {
					wwlog.Debug("api.getPower() node ipmi not configured")
					err = status.Wrap(fmt.Errorf("node ipmi not configured"), status.InvalidArgument)
					powerResults[0].Err = err // Set the error in the first PowerResult
					*output = powerResults    // Set *output on exit. important
					return err
				}

				// starting with powerResults[0] as the accumulator
				powerResults[0].BmcIp = node_.Ipmi.Ipaddr.String() // Set the BMC IP address in the PowerResult

				// Setup ipmitool call.
				ipmiCmd := bmc.TemplateStruct{
					IpmiConf: *node_.Ipmi,
				}

				// Make ipmitool call.
				result, err := ipmiCmd.PowerStatus()
				if err == nil {
					if result == "Chassis Power is on" {
						powerResults[0].Result = "on"
					} else if result == "Chassis Power is off" {
						powerResults[0].Result = "off"
					} else {
						powerResults[0].Result = "unknown"
						err = status.Wrap(fmt.Errorf("unknown power status: %s", result), status.Internal)
						powerResults[0].Err = err
						*output = powerResults // Set *output on exit. important
						return err
					}
					*output = powerResults // Set *output on exit. important
					wwlog.Debug("api.getPower() PASS running ipmi command, result, *output, err: %v, %v, %v", result, *output, err)
					return nil
				} else {
					wwlog.Debug("api.getPower() FAIL running ipmi command, result, err: %v, %v, %v", result, err)
					// IPMI is configured for the node. IPMI emulator is not
					// running or IPMI is not running on the BMC. This is an
					// error, but we don't want to return http 500.
					// http not available 503.
					err = status.Wrap(fmt.Errorf("unable to connect to node bmc"), status.Unavailable)
					powerResults[0].Err = err
					*output = powerResults // Set *output on exit. important
				}
				powerResults[0].Err = err
				*output = powerResults // Set *output on exit. important
				return err             // Odd to need this.
			}
		}
	})
	u.SetTitle("Get power status")
	u.SetDescription("Get power status for a node.")
	u.SetTags("Get Power")
	u.SetExpectedErrors(status.NotFound, status.Internal, status.InvalidArgument)
	return u
}

// validPowerStates is a list of valid power states that can be set for a node.
var validPowerStates = []string{"on", "off", "soft", "cycle", "reset"}

// setPower returns an interactor to set the power state of a node by its ID.
// At this time we ony support a single node. There is no expansion of the hostlist args or batching yet.
func setPower() usecase.Interactor {
	type setPowerInput struct {
		ID    string `path:"id" required:"true" description:"ID of node to get power status for"`
		State string `json:"state" required:"true" description:"Power state to set for the node (e.g., 'on', 'off', 'reset')"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input setPowerInput, output *[]PowerResult) error {
		wwlog.Debug("api.setPower(), ID: %s, State: %s", input.ID, input.State)
		powerResults := []PowerResult{}
		// accumulator
		powerResults = append(powerResults, PowerResult{
			NodeId: input.ID,
		})

		input.State = strings.ToLower(input.State)
		if !slices.Contains(validPowerStates, input.State) {
			err := status.Wrap(fmt.Errorf("invalid power state %s valid states are %v", input.State, validPowerStates), status.InvalidArgument)
			powerResults[0].Err = err
			*output = powerResults // Set *output on exit. important
		}

		if registry, err := node.New(); err != nil {
			wwlog.Debug("api.setPower() getting node registry: %v", err)
			powerResults[0].Err = err
			*output = powerResults // Set *output on exit. important
			return err
		} else {
			if node_, err := registry.GetNode(input.ID); err != nil {
				wwlog.Debug("api.setPower() node not found: %v", err)
				powerResults[0].Err = err
				*output = powerResults // Set *output on exit. important
				return status.Wrap(fmt.Errorf("node not found: %v (%v)", input.ID, err), status.NotFound)
			} else {
				if node_.Ipmi == nil || node_.Ipmi.Ipaddr.IsUnspecified() || node_.Ipmi.UserName == "" || node_.Ipmi.Password == "" {
					wwlog.Debug("api.getPower() node ipmi not configured")
					err = status.Wrap(fmt.Errorf("node ipmi not configured"), status.InvalidArgument)
					powerResults[0].Err = err
					*output = powerResults // Set *output on exit. important
					return err
				}

				powerResults[0].BmcIp = node_.Ipmi.Ipaddr.String() // Set the BMC IP address in the result now that we have it.

				ipmiCmd := bmc.TemplateStruct{
					IpmiConf: *node_.Ipmi,
				}

				// Set the power state based on the input.
				switch input.State {
				case "on":
					_, err = ipmiCmd.PowerOn()
				case "off":
					_, err = ipmiCmd.PowerOff()
				case "soft":
					_, err = ipmiCmd.PowerSoft()
				case "cycle":
					_, err = ipmiCmd.PowerCycle()
				case "reset":
					_, err = ipmiCmd.PowerReset()
				default:
					err = fmt.Errorf("unknown power state: %s", input.State)
					powerResults[0].Err = err
					*output = powerResults // Set *output on exit. important
					return status.Wrap(err, status.Unimplemented)
				}

				if err == nil {
					powerResults[0].Err = err // Set the result to the requested state
					*output = powerResults    // Set *output on exit. important
					return err
				} else {
					wwlog.Debug("api.setPower() FAIL running ipmi command: %v", err)
					err = status.Wrap(err, status.Unimplemented)
					powerResults[0].Err = err // Set the error in the first PowerResult
					*output = powerResults    // Set *output on exit. important
					return err
				}
			}
		}
	})
	u.SetTitle("Set power status")
	u.SetDescription("Set power status for a node.")
	u.SetTags("Set Power")
	u.SetExpectedErrors(
		status.InvalidArgument,
		status.NotFound,
		status.Unimplemented,
	)
	return u
}
