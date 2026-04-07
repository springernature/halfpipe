package cmds

var Quiet bool

var Input string

var Platform string

func init() {
	rootCmd.PersistentFlags().StringVarP(&Input, "input", "i", "", "Sets the halfpipe filename to be used")
	rootCmd.PersistentFlags().BoolVarP(&Quiet, "quiet", "q", false, "Suppress warnings")
	rootCmd.PersistentFlags().StringVarP(&Platform, "platform", "p", "", "Override the platform (actions or concourse)")
}
