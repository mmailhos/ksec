package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/fatih/color"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// kubeAPI returns a CoreV1Interface object to call Kubernetes API
func kubeAPI(kubeconfig clientcmd.ClientConfig) corev1.CoreV1Interface {
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		prExit(err)
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		prExit(err)
	}

	return clientset.CoreV1()
}

// secretData. Handy struct to convert secret data into a Slice to sort later
type secretDataType struct {
	Key   string
	Value string
}

// prExit prints the message and exit 1
func prExit(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

// filterFunc takes a *v1.Secret.Data map and a target string and filters the map to return a slice of KeyValue struct
type filterFunc func(map[string][]byte, string) []secretDataType

// sortFilter will take secretData input and return a sorted array of keys
func sortFilter(secretData map[string][]byte, targetData string) (secData []secretDataType) {
	for k, v := range secretData {
		if targetData == "" || strings.Contains(strings.ToLower(k), strings.ToLower(targetData)) {
			secData = append(secData, secretDataType{Key: k, Value: string(v)})
		}
	}
	if len(secData) == 0 {
		fmt.Fprintf(os.Stderr, "No data key found with %s\n", targetData)
	}
	sort.Slice(secData, func(i, j int) bool {
		return secData[i].Key < secData[j].Key
	})
	return
}

// printSecretYAML simply prints on STDOUT the data of a given secret in YAML format
func printSecretYAML(secret *v1.Secret, targetData string, filterF filterFunc, flagColor, flagMetadata bool) {
	keyColorF := color.New(color.FgBlue).SprintFunc()
	valueColorF := color.New(color.FgGreen).SprintFunc()

	printTitle := func(title string) {
		if flagColor {
			fmt.Println(keyColorF(title + ":"))
		} else {
			fmt.Println(title + ":")
		}
	}

	if flagMetadata {
		metadata := map[string]string{"Name": secret.Name, "Type": string(secret.Type), "Count": strconv.Itoa(len(secret.Data)), "Size": strconv.Itoa(secret.Size())}
		printTitle("metadata")
		for key := range metadata {
			if flagColor {
				fmt.Printf("  %s: %s\n", keyColorF(key), valueColorF(metadata[key]))
			} else {
				fmt.Printf("  %s: %s\n", key, metadata[key])
			}
		}
	}

	// We assume we do not need to have a subfield values if --metadata is not specified
	indent := ""
	if flagMetadata {
		printTitle("values")
		indent = "  "
	}

	for _, sd := range filterF(secret.Data, targetData) {
		if flagColor {
			fmt.Printf("%s%s: %s\n", indent, keyColorF(sd.Key), valueColorF(sd.Value))
		} else {
			fmt.Printf("%s%s: %s\n", indent, sd.Key, sd.Value)
		}
	}
}

// printSecretEnv simply prints on STDOUT the data of a given secret in bash-like Env format
func printSecretEnv(secret *v1.Secret, targetData string, filterF filterFunc, flagColor, flagMetadata bool) {
	keyColorF := color.New(color.FgBlue).SprintFunc()
	valueColorF := color.New(color.FgGreen).SprintFunc()

	if flagMetadata {
		metadata := map[string]string{"Name": secret.Name, "Type": string(secret.Type), "Count": strconv.Itoa(len(secret.Data)), "Size": strconv.Itoa(secret.Size())}
		for key := range metadata {
			envKey := "METADATA_" + strings.ToUpper(key)
			if flagColor {
				fmt.Printf("%s=%s\n", keyColorF(envKey), valueColorF(metadata[key]))
			} else {
				fmt.Printf("%s=%s\n", envKey, metadata[key])
			}
		}
	}
	for _, sd := range filterF(secret.Data, targetData) {
		if flagColor {
			fmt.Printf("%s=%s\n", keyColorF(sd.Key), valueColorF(sd.Value))
		} else {
			fmt.Printf("%s=%s\n", sd.Key, sd.Value)
		}
	}
}

// printSecretJSON do the final logic taken an array of secrets
// no support for color at this time
func printSecretJSON(secret *v1.Secret, targetData string, filterF filterFunc, flagMetadata bool) {
	type metadata struct {
		Name   string
		Type   string
		Length int
		Size   int
	}
	type outjson struct {
		Metadata metadata          `json:"metadata",omitempty`
		Values   map[string]string `json:"values"`
	}

	secs := outjson{Values: make(map[string]string)}
	for _, sd := range filterF(secret.Data, targetData) {
		secs.Values[sd.Key] = sd.Value
	}

	if flagMetadata {
		secs.Metadata = metadata{Name: secret.Name, Type: string(secret.Type), Length: len(secret.Data), Size: secret.Size()}
	}

	secretJSON, err := json.MarshalIndent(secs, "", "  ")
	if err != nil {
		prExit(err)
	}
	fmt.Fprintf(os.Stdout, "%s", secretJSON)
}

// getRelease returns the release number (or 0) of a given secret name
func getRelease(secretName string) (release int) {
	pattern := regexp.MustCompile(`(\d+)$`)
	release, _ = strconv.Atoi(pattern.FindString(secretName))
	return
}

// getSecrets returns the shortest list possible of secrets matching the target secret
func getSecrets(targetSecret, targetType string, secrets *v1.SecretList) (foundSecrets []*v1.Secret) {
	var maxVersionned int
	secretReleases := make(map[string]int)
	secretFamilies := make(map[string]*v1.Secret)

	for _, secret := range secrets.Items {
		if (targetType == "" || string(secret.Type) == targetType) && strings.Contains(secret.Name, targetSecret) {
			if matchVersioned, _ := regexp.MatchString("^"+targetSecret+"-[0-9]*$", secret.Name); matchVersioned {
				if release := getRelease(secret.Name); release >= maxVersionned {
					maxVersionned = release
					foundSecrets = []*v1.Secret{secret.DeepCopy()}
				}
			} else if len(foundSecrets) == 0 {
				release := getRelease(secret.Name)
				family := strings.TrimSuffix(secret.Name, strconv.Itoa(release))
				if release >= secretReleases[family] {
					secretReleases[family] = release
					secretFamilies[family] = secret.DeepCopy()
				}
			}
		}
	}
	// Flatten families if we haven't found any perfectly matching secret
	if len(foundSecrets) == 0 {
		for _, secret := range secretFamilies {
			foundSecrets = append(foundSecrets, secret.DeepCopy())
		}
	}
	return
}

func main() {
	var flagNamespace, flagLabel, flagField, flagType, flagOut, namespace, targetData string
	var flagColor, flagMetadata bool
	var err error

	flag.StringVar(&flagNamespace, "namespace", "", "namespace")
	flag.StringVar(&flagLabel, "label", "", "Label selector")
	flag.StringVar(&flagField, "field", "", "Field selector")
	flag.BoolVar(&flagColor, "color", false, "Use colors")
	flag.StringVar(&flagOut, "out", "env", "Output format (env, json, yaml)")
	flag.BoolVar(&flagMetadata, "metadata", false, "Print metadata of the found secret (Name, Type)")
	flag.StringVar(&flagType, "type", "", "Look for a specific secret type (ex: Opaque)")
	flag.Parse()

	if len(flag.Args()) == 0 {
		prExit(errors.New("Missing main argument"))
	}
	targetSecret := flag.Args()[0]

	if len(flag.Args()) > 1 {
		targetData = flag.Args()[1]
	}

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	if flagNamespace == "" {
		namespace, _, err = kubeconfig.Namespace()
		if err != nil {
			prExit(errors.New("Invalid namespace"))
		}
	}

	api := kubeAPI(kubeconfig)
	outSec := func(sec *v1.Secret) {
		if flagOut == "yaml" || flagOut == "yml" {
			printSecretYAML(sec, targetData, sortFilter, flagColor, flagMetadata)
		} else if flagOut == "env" {
			printSecretEnv(sec, targetData, sortFilter, flagColor, flagMetadata)
		} else if flagOut == "json" {
			printSecretJSON(sec, targetData, sortFilter, flagMetadata)
		} else {
			fmt.Fprintf(os.Stderr, "Unknown output format")
		}
	}

	// Attempt to directly get the secret (e.g. perfect match) to avoid unnecessary operations
	foundSecret, err := api.Secrets(namespace).Get(targetSecret, metav1.GetOptions{})
	if err == nil && (flagType == "" || flagType == string(foundSecret.Type)) {
		outSec(foundSecret)
		os.Exit(0)
	}

	// Retrieve all secrets and find potential matchs
	listOptions := metav1.ListOptions{LabelSelector: flagLabel, FieldSelector: flagField}
	secrets, err := api.Secrets(namespace).List(listOptions)
	if err != nil {
		prExit(err)
	}

	foundSecrets := getSecrets(targetSecret, flagType, secrets)
	switch len(foundSecrets) {
	case 0:
		fmt.Println("No secret found")
	case 1:
		outSec(foundSecrets[0])
	default:
		fmt.Println("Unable to determine the target. Try one of these:")
		for _, secret := range foundSecrets {
			fmt.Println(secret.Name)
		}
	}
}
