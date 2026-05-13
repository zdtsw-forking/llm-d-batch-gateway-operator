package controller

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	batchv1alpha1 "github.com/opendatahub-io/llm-d-batch-gateway-operator/api/v1alpha1"
)

var (
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
)

func TestMain(m *testing.M) {
	logf.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	if err := batchv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		logf.Log.Error(err, "failed to add scheme")
		os.Exit(1)
	}
	if err := gatewayv1beta1.Install(scheme.Scheme); err != nil {
		logf.Log.Error(err, "failed to add gateway-api scheme")
		os.Exit(1)
	}

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "config", "crd", "bases"),
			gatewayAPICRDPath(),
		},
		ErrorIfCRDPathMissing: true,
	}

	if dir := getFirstFoundEnvTestBinaryDir(); dir != "" {
		testEnv.BinaryAssetsDirectory = dir
	}

	var err error
	cfg, err = testEnv.Start()
	if err != nil {
		logf.Log.Error(err, "failed to start envtest")
		os.Exit(1)
	}

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	if err != nil {
		logf.Log.Error(err, "failed to create client")
		os.Exit(1)
	}

	code := m.Run()

	if err := testEnv.Stop(); err != nil {
		logf.Log.Error(err, "failed to stop envtest")
	}

	os.Exit(code)
}

// gatewayAPICRDPath returns the path to the gateway-api standard CRDs by
// asking the Go toolchain where the module is cached. This avoids hardcoding
// the version.
func gatewayAPICRDPath() string {
	cmd := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "sigs.k8s.io/gateway-api")
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if ee, ok := err.(*exec.ExitError); ok {
			stderr = string(ee.Stderr)
		}
		logf.Log.Error(err, "failed to locate sigs.k8s.io/gateway-api module", "stderr", stderr)
		return ""
	}
	return filepath.Join(strings.TrimSpace(string(out)), "config", "crd", "standard")
}

func getFirstFoundEnvTestBinaryDir() string {
	basePath := filepath.Join("..", "..", "bin", "k8s")
	entries, err := os.ReadDir(basePath)
	if err != nil {
		return ""
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(basePath, entry.Name())
		subEntries, err := os.ReadDir(subDir)
		if err != nil {
			continue
		}
		for _, sub := range subEntries {
			if sub.IsDir() {
				return filepath.Join(subDir, sub.Name())
			}
		}
	}
	return ""
}
