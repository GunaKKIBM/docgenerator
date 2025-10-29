package main


import (
	"fmt"
	"sigs.k8s.io/controller-tools/pkg/genall"
	"sigs.k8s.io/controller-tools/pkg/markers"
	"sigs.k8s.io/controller-tools/pkg/loader"
	"io"
	"strings"
	"golang.org/x/tools/go/packages"
	"path/filepath"
	"os"
	//"go/ast"
	//"reflect"
	//v1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)


var (
	RuleDefinition = markers.Must(markers.MakeDefinition("docgenerator:pod", markers.DescribesPackage, Rule{}))
)


type PodDocGenerator struct {
	DocName string
}

type PodSpec struct {
	Image string `marker:",optional"`
	Version string `marker:",optional"`
}

type Rule struct {
	//PodSpec markers.RawArguments `marker:",optional"`
	Scenario string `marker:",optional"`
	SuccessStates []string `marker:",optional"`
	FailedStates []string `marker:",optional"`
	FailureReason string `marker:",optional"`
	HasInitContainer string `marker:",optional"`
	HasVolume string `marker:",optional"`
}


func (PodDocGenerator) RegisterMarkers(into *markers.Registry) error {
	if err := into.Register(RuleDefinition); err != nil {
		return err
	}
	//into.AddHelp(RuleDefinition, Rule{}.Help())
	return nil
}

func (pd PodDocGenerator) Generate(ctx *genall.GenerationContext) error {

	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("# %s\n\n", "POD CONDITION TESTS"))
	template := "### Scenario: %s\n\n**PodSpec**:\n```sh\n%s\n```\n\n- For the above pod spec, pod successfully transitions to **%s** states\n"

	templateFailure := "- Fails to transition to **%s** states\n\n**Reason for failure**: %s\n"


	for _, root := range ctx.Roots {
		markerSet, err := markers.PackageMarkers(ctx.Collector, root)
		fmt.Println(markerSet)
		if err != nil {
			fmt.Println(err)
		}

		for _, value := range markerSet[RuleDefinition.Name] {
			rule := value.(Rule)
			var opvalue string
			if len(rule.FailedStates) != 0 {
				opvalue = fmt.Sprintf(template + templateFailure, rule.Scenario, podSpec(rule.HasInitContainer, rule.HasVolume), strings.Join(rule.SuccessStates, " -> "), strings.Join(rule.FailedStates, " ,"), rule.FailureReason)
			} else {
				opvalue = fmt.Sprintf(template + "\n\n", rule.Scenario, podSpec(rule.HasInitContainer, rule.HasVolume), strings.Join(rule.SuccessStates, " -> "))
			}
			
			builder.WriteString(opvalue)
			
        	err := os.WriteFile(pd.DocName, []byte(builder.String()), 0644)
        	if err != nil {
                fmt.Printf("Error writing file: %v\n", err)
			}
		}
	}
	fmt.Println(pd.DocName)
	return nil
}


type OutputDoc struct {
	// Config points to the directory to which to write configuration.
	Config genall.OutputToDirectory
	// Code overrides the directory in which to write new code (defaults to where the existing code lives).
	Code genall.OutputToDirectory `marker:",optional"`
}

func (od OutputDoc) Open(pkg *loader.Package, itemPath string) (io.WriteCloser, error) {
		if pkg == nil {
		return od.Config.Open(pkg, itemPath)
	}

	if od.Code != "" {
		return od.Code.Open(pkg, itemPath)
	}

	if len(pkg.CompiledGoFiles) == 0 {
		return nil, fmt.Errorf("cannot output to a package with no path on disk")
	}
	outDir := filepath.Dir(pkg.CompiledGoFiles[0])
	if _, err := os.Stat(outDir); os.IsNotExist(err) {
		if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
			return nil, err
		}
	}

	outPath := filepath.Join(outDir, itemPath)
	return os.Create(outPath)
}

func main() {
	var allGenerators = map[string]genall.Generator{
		"pod": PodDocGenerator{},
	}

	var allOutputRules = map[string]genall.OutputRule{
		"doc":       OutputDoc{},
	}

	var optionsRegistry = &markers.Registry{}

		for genName, gen := range allGenerators {
		// make the generator options marker itself
		defn := markers.Must(markers.MakeDefinition(genName, markers.DescribesPackage, gen))
		if err := optionsRegistry.Register(defn); err != nil {
			panic(err)
		}
		if helpGiver, hasHelp := gen.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(defn, help)
			}
		}

		// make per-generation output rule markers
		for ruleName, rule := range allOutputRules {
			ruleMarker := markers.Must(markers.MakeDefinition(fmt.Sprintf("output:%s:%s", genName, ruleName), markers.DescribesPackage, rule))
			if err := optionsRegistry.Register(ruleMarker); err != nil {
				panic(err)
			}
			if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
				if help := helpGiver.Help(); help != nil {
					optionsRegistry.AddHelp(ruleMarker, help)
				}
			}
		}
	}

	for ruleName, rule := range allOutputRules {
		ruleMarker := markers.Must(markers.MakeDefinition("output:"+ruleName, markers.DescribesPackage, rule))
		if err := optionsRegistry.Register(ruleMarker); err != nil {
			panic(err)
		}
		if helpGiver, hasHelp := rule.(genall.HasHelp); hasHelp {
			if help := helpGiver.Help(); help != nil {
				optionsRegistry.AddHelp(ruleMarker, help)
			}
		}
	}

	if err := genall.RegisterOptionsMarkers(optionsRegistry); err != nil {
		panic(err)
	}

	buildTags := []string{"load-build-tags"}

	tagsFlag := fmt.Sprintf("-tags=%s", strings.Join(buildTags, ","))

	rawOpts := []string{"pod:docName=PODStates.md", "paths=\"./...\"", "output:pod:doc:config=docs"}

	rt, err := genall.FromOptionsWithConfig(&packages.Config{BuildFlags: []string{tagsFlag}}, optionsRegistry, rawOpts)

	if err != nil {
		fmt.Println(err)
		return
	}

	if hadErrs := rt.Run(); hadErrs {
		fmt.Errorf("not all generators ran successfully")
	}

}
/*
func podSpec(addInitContainer bool) string {
	p := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "<Podname>",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "<containerName>",
					Image: "<image>",
					Args:  []string{"test-webserver"},
				},
			},
		},
	}
	if addInitContainer {
		p.Spec.InitContainers = []v1.Container{
			{
				Name:    "<initContainerName>",
				Image:   "image2",
				Command: []string{"sh", "-c", "sleep 5s"},
			},
		}
	}

	p.Spec.Volumes = []v1.Volume{
			{
				Name: "cm",
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{Name: "does-not-exist"},
					},
				},
			},
		}
		p.Spec.Containers[0].VolumeMounts = []v1.VolumeMount{
			{
				Name:      "cm",
				MountPath: "/config",
			},
		}

	podSpecMarkdown := 
"v1.PodSpec{\n" +
"    Containers: []v1.Container{\n" +
"        {\n" +
"            Name:  \"<containerName>\",\n" +
"            Image: \"<image>\",\n" +
"            Args:  []string{\"test-webserver\"},\n" +
"        },\n" +
"    },\n" +
"    InitContainers: []v1.Container{\n" +
"        {\n" +
"            Name:    \"<initContainerName>\",\n" +
"            Image:   \"image2\",\n" +
"            Command: []string{\"sh\", \"-c\", \"sleep 5s\"},\n" +
"        },\n" +
"    },\n" +
"    Volumes: []v1.Volume{\n" +
"        {\n" +
"            Name: \"cm\",\n" +
"            VolumeSource: v1.VolumeSource{\n" +
"                ConfigMap: &v1.ConfigMapVolumeSource{\n" +
"                    LocalObjectReference: v1.LocalObjectReference{Name: \"does-not-exist\"},\n" +
"                },\n" +
"            },\n" +
"        },\n" +
"    },\n" +
"}\n"

	return podSpecMarkdown
}
	*/


func podSpec(addInitContainer string, addVolume string) string {

	podSpecMarkdownBase := 
"v1.PodSpec{\n" +
"    Containers: []v1.Container{\n" +
"        {\n" +
"            Name:  \"<containerName>\",\n" +
"            Image: \"<image>\",\n" +
"            Args:  []string{\"test-webserver\"},\n" +
"        },\n" +
"    },\n"


	podSpecMarkdownInitContainers := 
"    InitContainers: []v1.Container{\n" +
"        {\n" +
"            Name:    \"<initContainerName>\",\n" +
"            Image:   \"image2\",\n" +
"            Command: []string{\"sh\", \"-c\", \"sleep 5s\"},\n" +
"        },\n" +
"    },\n" 


	podSpecMarkdownVolumes := 
"    Volumes: []v1.Volume{\n" +
"        {\n" +
"            Name: \"cm\",\n" +
"            VolumeSource: v1.VolumeSource{\n" +
"                ConfigMap: &v1.ConfigMapVolumeSource{\n" +
"                    LocalObjectReference: v1.LocalObjectReference{Name: \"does-not-exist\"},\n" +
"                },\n" +
"            },\n" +
"        },\n" +
"    },\n" +
"}\n"

	podSpec := podSpecMarkdownBase
	if addInitContainer == "true" {
		podSpec = podSpec + podSpecMarkdownInitContainers
		
	}

	if addVolume  == "true" {
		podSpec = podSpec + podSpecMarkdownVolumes
	}


	return podSpec
}