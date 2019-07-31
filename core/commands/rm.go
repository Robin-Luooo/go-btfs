package commands

import (
	"fmt"

	"github.com/TRON-US/go-btfs/core/commands/cmdenv"
	"github.com/TRON-US/go-btfs/namesys/resolve"
	cmds "github.com/ipfs/go-ipfs-cmds"
	path2 "github.com/ipfs/go-path"
	"github.com/ipfs/interface-go-ipfs-core/options"
	"github.com/ipfs/interface-go-ipfs-core/path"
)

var RmCmd = &cmds.Command{
	Helptext: cmds.HelpText{
		Tagline:          "Remove a file or directory from a local btfs node.",
		ShortDescription: `Removes contents of <hash> from a local btfs node.`,
	},

	Arguments: []cmds.Argument{
		cmds.StringArg("hash", true, true, "The hash of the file to be removed from local btfs node.").EnableStdin(),
	},
	Run: func(req *cmds.Request, res cmds.ResponseEmitter, env cmds.Environment) error {
		n, err := cmdenv.GetNode(env)
		if err != nil {
			return err
		}

		api, err := cmdenv.GetApi(env, req)
		if err != nil {
			return err
		}

		// Since we are removing a file, we need to set recursive flag to true
		recursive := true

		if err := req.ParseBodyArgs(); err != nil {
			return err
		}

		// Remove pins recursively
		enc, err := cmdenv.GetCidEncoder(req)
		if err != nil {
			return err
		}

		pins := make([]string, 0, len(req.Arguments))
		for _, b := range req.Arguments {
			rp, err := api.ResolvePath(req.Context, path.New(b))
			if err != nil {
				return err
			}

			id := enc.Encode(rp.Cid())
			pins = append(pins, id)
			if err := api.Pin().Rm(req.Context, rp, options.Pin.RmRecursive(recursive)); err != nil {
				return err
			}
		}

		// Surgincal approach
		p, err := path2.ParsePath(req.Arguments[0])
		if err != nil {
			return err
		}

		object, err := resolve.Resolve(req.Context, n.Namesys, n.Resolver, p)
		if err != nil {
			return err
		}

		// rm all child links
		for _, cid := range object.Links() {
			if err := n.Blockstore.DeleteBlock(cid.Cid); err == nil {
				fmt.Printf("Removed %s\n", cid.Cid.Hash().B58String())
			}
		}

		// rm parent node
		if err := n.Blockstore.DeleteBlock(object.Cid()); err == nil {
			fmt.Printf("Removed %s\n", object.Cid().Hash().B58String())
		}

		return nil
	},
}