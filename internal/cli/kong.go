package cli

type CLI struct {
	List     ListCmd     `cmd:"" help:"List eligible PIM role assignments"`
	Status   StatusCmd   `cmd:"" help:"List active PIM role assignments"`
	Activate ActivateCmd `cmd:"" help:"Activate an eligible PIM role assignment"`
}
