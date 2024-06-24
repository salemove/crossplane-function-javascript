# Building a function with external dependencies

This example demonstrates how to package external dependencies into a
single text file and package it into a Crossplane Composition resource.

The function source code is located in [`src/index.js`](./src/index.js) file,
and it uses NPM package [YAML](https://www.npmjs.com/package/yaml) as an external
dependency.

Eternal dependencies are bundled into a single file by [esbuild] and then transpiled
with [Babel] to ES5 syntax supported by [Goja]. The resulting source code is used to
build the [`composition.yaml`](./composition.yaml) file.

By default, the function server also _transpiles_ the input code into ES 5.1 syntax using
[Babel], but since the input code is already ES 5.1, `.spec.source.transpile` is set to `false`.
Setting this flag for large bundles can significantly improve performance.

[esbuild]: https://esbuild.github.io/
[Goja]: https://github.com/dop251/goja
[Babel]: https://babeljs.io/

## Running this example

Make sure you have Node.js installed.

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
# ...function results
```

Stop the function running in background:
```shell
$ fg
# Press Ctrl-C
^C
```
