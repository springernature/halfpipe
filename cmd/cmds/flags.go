package cmds

var Quiet bool

var Input string

func init() {
	rootCmd.PersistentFlags().StringVarP(&Input, "input", "i", "", "Sets the halfpipe filename to be used")

	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "suppress warnings")
}
