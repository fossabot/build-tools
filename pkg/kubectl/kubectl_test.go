package kubectl

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"gitlab.com/sparetimecoders/build-tools/pkg/config"
	"io"
	"io/ioutil"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"os"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	k := New(&config.Environment{Context: "missing", Namespace: "dev", Name: "dummy"})

	assert.Equal(t, "missing", k.(*kubectl).context)
	assert.Equal(t, "dev", k.(*kubectl).namespace)
}

func TestNew_NoNamespace(t *testing.T) {
	k := New(&config.Environment{Context: "missing", Namespace: "", Name: "dummy"})

	assert.Equal(t, "missing", k.(*kubectl).context)
	assert.Equal(t, "default", k.(*kubectl).namespace)
}

func TestKubectl_Apply(t *testing.T) {
	calls = [][]string{}
	newKubectlCmd = mockCmd
	tempDir, _ := ioutil.TempDir(os.TempDir(), "build-tools")

	k := &kubectl{context: "missing", namespace: "default", environment: &config.Environment{Context: "missing", Namespace: "default", Name: "dummy"}, tempDir: tempDir}

	err := k.Apply("")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"apply", "--context", "missing", "--namespace", "default", "--file", fmt.Sprintf("%s/content.yaml", tempDir), "--timout", "0s", "--show-events", "false", "--selector", ""}, calls[0])
}

func TestKubectl_UnableToCreateTempDir(t *testing.T) {
	newKubectlCmd = mockCmd

	k := &kubectl{context: "missing", namespace: "default", environment: &config.Environment{}, tempDir: "/missing"}

	err := k.Apply("")
	assert.EqualError(t, err, "open /missing/content.yaml: no such file or directory")
}

func TestKubectl_Environment(t *testing.T) {
	env := &config.Environment{Context: "missing", Namespace: "default", Name: "dummy"}
	k := New(env)

	assert.Equal(t, env, k.Environment())
}

func TestKubectl_DeploymentExistsTrue(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.DeploymentExists("image")
	assert.True(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"get", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "false", "--selector", ""}, calls[0])
}

func TestKubectl_DeploymentExistsFalse(t *testing.T) {
	calls = [][]string{}
	e := "deployment not found"
	cmdError = &e
	newKubectlCmd = mockCmd

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.DeploymentExists("image")
	assert.False(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"get", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "false", "--selector", ""}, calls[0])
}

func TestKubectl_RolloutStatusSuccess(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.RolloutStatus("image")
	assert.True(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"rollout", "status", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "1m0s", "--show-events", "false", "--selector", ""}, calls[0])
}

func TestKubectl_RolloutStatusFailure(t *testing.T) {
	calls = [][]string{}
	e := "rollout failed"
	cmdError = &e
	newKubectlCmd = mockCmd

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.RolloutStatus("image")
	assert.False(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"rollout", "status", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "1m0s", "--show-events", "false", "--selector", ""}, calls[0])
}

func TestKubectl_RolloutStatusFatal(t *testing.T) {
	calls = [][]string{}
	e := "rollout failed"
	cmdError = &e
	fatal = true
	defer func() { fatal = false }()

	newKubectlCmd = mockCmd

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.RolloutStatus("image")
	assert.False(t, result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"rollout", "status", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "1m0s", "--show-events", "false", "--selector", ""}, calls[0])
}

func TestKubectl_DeploymentEvents_Error(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := "deployment not found"
	cmdError = &e

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.DeploymentEvents("image")
	assert.Equal(t, "deployment not found", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "true", "--selector", ""}, calls[0])
}

func TestKubectl_DeploymentEvents_NoEvents(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Name:               gpe-core
Namespace:          default
Events:          <none>
`
	events = &e

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.DeploymentEvents("image")
	assert.Equal(t, "", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "true", "--selector", ""}, calls[0])
}

func TestKubectl_DeploymentEvents_SomeEvents(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Name:               gpe-core
Namespace:          default
Events:
  Type    Reason             Age   From                   Message
  ----    ------             ----  ----                   -------
  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled up replica set gpe-core-5cb459ff7d to 1
  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled down replica set gpe-core-7fc44679dc to 0
  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled up replica set gpe-core-c8798ff88 to 1
  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled down replica set gpe-core-5cb459ff7d to 0
`
	events = &e

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.DeploymentEvents("image")
	assert.Equal(t, "Events:\n  Type    Reason             Age   From                   Message\n  ----    ------             ----  ----                   -------\n  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled up replica set gpe-core-5cb459ff7d to 1\n  Normal  ScalingReplicaSet  9m    deployment-controller  Scaled down replica set gpe-core-7fc44679dc to 0\n  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled up replica set gpe-core-c8798ff88 to 1\n  Normal  ScalingReplicaSet  61s   deployment-controller  Scaled down replica set gpe-core-5cb459ff7d to 0\n", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "deployment", "image", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "true", "--selector", ""}, calls[0])
}

func TestKubectl_PodEvents_Error(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := "pod not found"
	cmdError = &e

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.PodEvents("image")
	assert.Equal(t, "pod not found", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "pods", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "true", "--selector", "app=image"}, calls[0])
}

func TestKubectl_PodEvents_NoEvents(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Name:               gpe-core
Namespace:          default
Events:          <none>
`
	events = &e

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.PodEvents("image")
	assert.Equal(t, "", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "pods", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "true", "--selector", "app=image"}, calls[0])
}

func TestKubectl_PodEvents_SomeEvents(t *testing.T) {
	calls = [][]string{}
	cmdError = nil
	newKubectlCmd = mockCmd
	e := `
Events:
  Type     Reason     Age                From                                                 Message
  ----     ------     ----               ----                                                 -------
  Normal   Scheduled  61s                default-scheduler                                    Successfully assigned dev/gpe-core-c8798ff88-674tr to some-ip-somewhere
  Normal   Pulling    10s (x4 over 60s)  kubelet, some-ip-somewhere                           pulling image "quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c"
  Normal   Pulled     9s (x4 over 57s)   kubelet, some-ip-somewhere                           Successfully pulled image "quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c"
  Normal   Created    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Created container
  Normal   Started    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Started container
  Warning  BackOff    8s (x5 over 54s)   kubelet, some-ip-somewhere                           Back-off restarting failed container`
	events = &e

	k := New(&config.Environment{Context: "missing", Namespace: "default", Name: "dummy"})

	result := k.PodEvents("image")
	assert.Equal(t, "Events:\n  Type     Reason     Age                From                                                 Message\n  ----     ------     ----               ----                                                 -------\n  Normal   Scheduled  61s                default-scheduler                                    Successfully assigned dev/gpe-core-c8798ff88-674tr to some-ip-somewhere\n  Normal   Pulling    10s (x4 over 60s)  kubelet, some-ip-somewhere                           pulling image \"quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c\"\n  Normal   Pulled     9s (x4 over 57s)   kubelet, some-ip-somewhere                           Successfully pulled image \"quay.io/somewhere/gpe-core:9cdb0243e82b9bfdf037627d9d59cbfcbf55406c\"\n  Normal   Created    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Created container\n  Normal   Started    8s (x4 over 57s)   kubelet, some-ip-somewhere                           Started container\n  Warning  BackOff    8s (x5 over 54s)   kubelet, some-ip-somewhere                           Back-off restarting failed container\n", result)
	assert.Equal(t, 1, len(calls))
	assert.Equal(t, []string{"describe", "pods", "--context", "missing", "--namespace", "default", "--file", "", "--timout", "0s", "--show-events", "true", "--selector", "app=image"}, calls[0])
}

var calls [][]string
var cmdError *string
var events *string
var fatal = false

func mockCmd(in io.Reader, out, err io.Writer) *cobra.Command {
	var ctx, ns, file *string
	var timeout *time.Duration
	var showEvents *bool
	var selector *string

	cmd := cobra.Command{
		Use: "kubectl",
		Args: func(cmd *cobra.Command, args []string) error {
			calls = append(calls, append(args, "--context", *ctx, "--namespace", *ns, "--file", *file, "--timout", timeout.String(), "--show-events", fmt.Sprintf("%v", *showEvents), "--selector", *selector))
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if fatal {
				util.CheckErr(errors.New(*cmdError))
			}
			if cmdError != nil {
				return errors.New(*cmdError)
			}
			if events != nil {
				_, _ = out.Write([]byte(*events))
			}
			return nil
		},
	}

	ctx = cmd.Flags().StringP("context", "c", "", "")
	ns = cmd.Flags().StringP("namespace", "n", "", "")
	file = cmd.Flags().StringP("filename", "f", "", "")
	timeout = cmd.Flags().DurationP("timeout", "t", 0*time.Second, "")
	showEvents = cmd.Flags().BoolP("show-events", "", false, "")
	selector = cmd.Flags().StringP("selector", "l", "", "")

	return &cmd
}