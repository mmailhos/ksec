package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/fatih/color"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// kubeAPI returns a CoreV1Interface object to call Kubernetes API
func kubeAPI(kubeconfig clientcmd.ClientConfig) corev1.CoreV1Interface {
	restconfig, err := kubeconfig.ClientConfig()
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(restconfig)
	if err != nil {
		panic(err)
	}

	return clientset.CoreV1()
}

// printResult do the final logic taken an array of secrets
func printResult(foundSecrets []*v1.Secret, flagColor bool) {
	if len(foundSecrets) == 0 {
		fmt.Println("No secret found")
	} else if len(foundSecrets) == 1 {
		for _, secret := range foundSecrets {
			printSecret(secret, flagColor)
		}
	} else {
		fmt.Println("Unable to determine the target. Try one of these:")
		for _, secret := range foundSecrets {
			fmt.Println(secret.Name)
		}
	}
}

// printSecret simply prints on STDOUT the data of a given secret
func printSecret(secret *v1.Secret, flagColor bool) {
	fmt.Printf("%s (%s)\n", secret.Name, secret.Type)
	for k, v := range secret.Data {
		if !flagColor {
			color.New(color.FgBlue).Printf(k)
			fmt.Printf(": ")
			color.New(color.FgGreen).Println(string(v))
		} else {
			fmt.Printf("%s: %s\n", k, string(v))
		}
	}
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
	var flagNamespace, flagLabel, flagField, flagType, namespace string
	var flagColor bool
	var err error

	flag.StringVar(&flagNamespace, "namespace", "", "namespace")
	flag.StringVar(&flagLabel, "label", "", "Label selector")
	flag.StringVar(&flagField, "field", "", "Field selector")
	flag.BoolVar(&flagColor, "no-color", false, "Do not use colors")
	flag.StringVar(&flagType, "type", "", "Look for a specific secret type (ex: Opaque)")
	flag.Parse()

	if len(flag.Args()) == 0 {
		panic(errors.New("missing argument secret"))
	}
	targetSecret := flag.Args()[0]

	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{},
	)
	if flagNamespace == "" {
		namespace, _, err = kubeconfig.Namespace()
		if err != nil {
			panic(err)
		}
	}

	api := kubeAPI(kubeconfig)

	// Attempt to directly get the secret (e.g. perfect match) to avoid unnecessary operations
	foundSecret, err := api.Secrets(namespace).Get(targetSecret, metav1.GetOptions{})
	if err == nil && (flagType == "" || flagType == string(foundSecret.Type)) {
		printSecret(foundSecret, flagColor)
		os.Exit(0)
	}

	// Retrieve all secrets and find potential matchs
	listOptions := metav1.ListOptions{LabelSelector: flagLabel, FieldSelector: flagField}
	secrets, err := api.Secrets(namespace).List(listOptions)
	if err != nil {
		panic(err)
	}
	printResult(getSecrets(targetSecret, flagType, secrets), flagColor)
}
