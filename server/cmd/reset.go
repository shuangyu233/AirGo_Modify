package cmd

import (
	"github.com/shuangyu233/AirGo_Modify/app"
	"github.com/shuangyu233/AirGo_Modify/service"
	"github.com/spf13/cobra"
)

func init() {
	resetCmd.Flags().StringVar(&startConfigPath, "config", "config.yaml", "config.yaml directory to read from")
	resetCmd.Flags().Bool("resetAdmin", false, "reset administrator account and password")
	resetCmd.MarkFlagsOneRequired("resetAdmin")
}

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "reset resources",
	Run: func(cmd *cobra.Command, args []string) {
		reset()
	},
	//Args: cobra.MaximumNArgs(1),
}

func reset() {
	newApp := app.NewApp()

	newApp.InitConfig(startConfigPath)

	newApp.InitLogrus()

	newApp.InitLocalCache()

	newApp.InitRouter()

	newApp.ConnectDatabase()

	service.AdminUserSvc.ResetAdminPassword()

}
