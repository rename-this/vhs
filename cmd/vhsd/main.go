package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json"
	"github.com/spf13/cobra"

	cmd2 "github.com/rename-this/vhs/cmd"
	"github.com/rename-this/vhs/session"
)

func main() {
	newRootCmd().Execute()
}

type VhsService struct {
	Cfg *session.Config
}

type StartFlowReply struct {
	Placeholder string // TODO: flesh this out
}

func (v *VhsService) StartFlow(r *http.Request, args *session.FlowConfig, reply *StartFlowReply) error {
	reply.Placeholder = "Hello vhsd"
	return nil
}

func newRootCmd() *cobra.Command {
	var (
		cmd = &cobra.Command{
			Short: "A daemon for running VHS flows.",
		}

		cfg = &session.Config{}
	)

	cmd2.SetConfigFlags(cmd, cfg)
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if err := root(cfg); err != nil {
			fmt.Printf("failed to initialize vhsd: %v", err)
		}
	}
	return cmd
}

func root(cfg *session.Config) error {
	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	if err := s.RegisterService(new(VhsService), ""); err != nil {
		return err
	}
	http.Handle("/rpc", s)
	if err := http.ListenAndServe(":10000", nil); err != nil {
		return err
	}
	return nil
}
