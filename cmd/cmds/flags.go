package cmds

var Quiet bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "suppress warnings")
}
