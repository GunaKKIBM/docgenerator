# POD CONDITION TESTS

### Scenario: Pod fails to mount in case of missing volumes

**PodSpec**:
```sh
v1.PodSpec{
    Containers: []v1.Container{
        {
            Name:  "<containerName>",
            Image: "<image>",
            Args:  []string{"test-webserver"},
        },
    },
    InitContainers: []v1.Container{
        {
            Name:    "<initContainerName>",
            Image:   "image2",
            Command: []string{"sh", "-c", "sleep 5s"},
        },
    },
    Volumes: []v1.Volume{
        {
            Name: "cm",
            VolumeSource: v1.VolumeSource{
                ConfigMap: &v1.ConfigMapVolumeSource{
                    LocalObjectReference: v1.LocalObjectReference{Name: "does-not-exist"},
                },
            },
        },
    },
}

```

- For the above pod spec, pod successfully transitions to **PodScheduled -> PodInitialized** states
- Fails to transition to **ContainersReady ,PodReady** states

**Reason for failure**: MountVolume.SetUp failed for volume config-volume : configmap my-config not found
