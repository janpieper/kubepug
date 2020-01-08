package main

import (
	"os"

	"github.com/rikatz/kubepug/pkg/kubepug"
	"github.com/spf13/cobra"

	"k8s.io/cli-runtime/pkg/genericclioptions"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	kubernetesConfigFlags *genericclioptions.ConfigFlags
)

var (
	k8sVersion      string
	forceDownload   bool
	apiWalk         bool
	swaggerDir      string
	showDescription bool

	rootCmd = &cobra.Command{
		Use:          "kubepug",
		SilenceUsage: true,
		Short:        "Verifies objects in deprecated APIs against a swagger.json from Kubernetes. It must run with a Cluster Admin account as it checks all the Cluster Objects",
		Example:      "  kubepug",
		Args:         cobra.MinimumNArgs(0),
		RunE:         runPug,
	}
)
var KubernetesAPIs map[string]kubepug.KubeAPI

func runPug(cmd *cobra.Command, args []string) error {

	swaggerfile, err := kubepug.DownloadSwaggerFile(k8sVersion, swaggerDir, forceDownload)

	if err != nil {
		return err
	}

	config, err := kubernetesConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	KubernetesAPIs = make(map[string]kubepug.KubeAPI)
	KubernetesAPIs = kubepug.PopulateKubeAPIMap(config, swaggerfile)

	// First lets List all the deprecated APIs
	kubepug.ListDeprecated(config, KubernetesAPIs, showDescription)

	if apiWalk {
		kubepug.WalkObjects(config, KubernetesAPIs)
	}

	return nil

}

func init() {
	kubernetesConfigFlags = genericclioptions.NewConfigFlags(true)
	kubernetesConfigFlags.AddFlags(rootCmd.Flags())
	rootCmd.Flags().MarkHidden("as")
	rootCmd.Flags().MarkHidden("as-group")
	rootCmd.Flags().MarkHidden("cache-dir")
	rootCmd.Flags().MarkHidden("certificate-authority")
	rootCmd.Flags().MarkHidden("client-certificate")
	rootCmd.Flags().MarkHidden("client-key")
	rootCmd.Flags().MarkHidden("insecure-skip-tls-verify")
	rootCmd.Flags().MarkHidden("namespace")
	rootCmd.Flags().MarkHidden("request-timeout")
	rootCmd.Flags().MarkHidden("server")
	rootCmd.Flags().MarkHidden("token")
	rootCmd.Flags().MarkHidden("user")

	rootCmd.PersistentFlags().BoolVar(&apiWalk, "api-walk", true, "Wether to walk in the whole API, checking if all objects type still exists in the current swagger.json. May be IO intensive to APIServer. Defaults to true")
	rootCmd.PersistentFlags().BoolVar(&showDescription, "description", true, "Wether to show the description of the deprecated object. The description may contain the solution for the deprecation. Defaults to true")
	rootCmd.PersistentFlags().StringVar(&k8sVersion, "k8s-version", "master", "Which kubernetes release version (https://github.com/kubernetes/kubernetes/releases) should be used to validate objects. Defaults to master")
	rootCmd.PersistentFlags().StringVar(&swaggerDir, "swagger-dir", "", "Where to keep swagger.json downloaded file. Defaults to current directory")
	rootCmd.PersistentFlags().BoolVar(&forceDownload, "force-download", false, "Wether to force the download of a new swagger.json file even if one exists. Defaults to false")

}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}