# JavaScript Function Example

You can run your function locally and test it using `crossplane beta render`
with these example manifests.

Run the function locally in the background:
```shell
$ make run &
```

Then call it with example manifests:
```shell
$ make render
---
apiVersion: example.crossplane.io/v1
kind: XR
metadata:
  name: example-xr
---
# ... results
```

Stop the function running in background:
```shell
$ fg
# Press Ctrl-C
^C
```
