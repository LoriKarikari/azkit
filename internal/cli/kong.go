package cli

type CLI struct {
	List     ListCmd     `cmd:"" help:"List eligible PIM role assignments"`
	Activate ActivateCmd `cmd:"" help:"Activate an eligible PIM role assignment"`
}
