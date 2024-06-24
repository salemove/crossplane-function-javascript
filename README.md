# function-javascript

[![CI](https://github.com/salemove/crossplane-function-javascript/actions/workflows/ci.yml/badge.svg)](https://github.com/salemove/crossplane-function-javascript/actions/workflows/ci.yml)

A function for writing [composition functions][functions] in ECMAScript/JavaScript.

Here's an example:
```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: function-javascript
spec:
  compositeTypeRef:
    apiVersion: example.crossplane.io/v1
    kind: XR
  mode: Pipeline
  pipeline:
  - step: run-the-template
    functionRef:
      name: function-javascript
    input:
      apiVersion: javascript.fn.crossplane.io/v1beta1
      kind: Input
      spec:
        source:
          inline: |
            export default (req, rsp) => {
              const composite = req.observed.composite.resource;

              rsp.setDesiredComposedResource('bucket', {
                apiVersion: 'example.org/v1alpha1',
                kind: 'Bucket',
                metadata: {
                  spec: {
                    region: composite.spec.region
                  }
                }
              });

              if (req.observed.resources?.bucket) {
                // Expose some connection details, get value from a resource generated within this function.
                // The function expects Base64-encoded strings. Use "btoa" function to encode plain strings.
                // ConnectionDetails from observed resources are already Base64-encoded.
                rsp.setConnectionDetails({
                  bucketName: btoa(req.observed.resources.bucket.resource.metadata.name) 
                });

                // patch composite resource status
                rsp.updateCompositeStatus({ bucketName: req.observed.resources.bucket.resource.metadata.name });
              }
            };
  - step: automatically-detect-ready-composed-resources
    functionRef:
       name: function-auto-ready
```

## Install the JavaScript function to Cluster

```shell
cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1beta1
kind: Function
metadata:
  name: function-javascript
spec:
  package: docker.io/salemove/crossplane-function-javascript:v0.2.0
EOF
```

## Using this function

At the moment, the function code can only be specified through `Inline` source.

The JavaScript runtime is based on [Goja][goja] and expects the program to export
a default function. The exported function is called with 2 arguments:
* `request` - a [`RunFunctionRequest`][req] object converted into a nested plain map.
  This means that you can access the composite resource, any composed resources, and
  the function pipeline context using notation like:
  * `request.observed.composite.resource.metadata.name`
  * `request.observed.resources.mywidget.resource.status.widgets`
  * `request.observed.resources.mywidget.connectionDetails`
  * `request.context["apiextensions.crossplane.io/environment"]`
  * `request.context["apiextensions.crossplane.io/extra-resources"].mywidget[0]`
* `response` - an object through which you can manipulate the function [response][resp].
   The object has the following methods:
   * `response.setDesiredComposedResource(name, properties)` - set the desired composed
     resource for the current function. The resource properties are passed as plain map.

     To mark a desired resource as ready, use the `javascript.fn.crossplane.io/ready` annotation:
     ```javascript
     export default function (req, rsp) {
       rsp.setDesiredCompositeResource('bucket', {
         apiVersion: 'example.org/v1',
         kind: 'Bucket',
         metadata: {
           annotations: { 'javascript.fn.crossplane.io/ready': 'True' }
         },
         spec: {
           // ...skipped for brevity
         }
       });
     }
     ```
   * `response.setConnectionDetails(details)` - sets the desired composite resource
     connection details.

     Connection details values must be Base64-encoded, use function `btoa` to encode
     plain strings to Base64.

     Connection details from other observed resources are already Base64-encoded, so
     you can pass their values to `setConnectionDetails` function as is:
     ```javascript
     export default function (req, rsp) {
       // ...skip for brevity
       const username = req.observed.resources.user.connectionDetails.username;
       const host = "localhost";

       rsp.setConnectionDetails({
         username,
         host: btoa(host)
       });
     }
     ```
   * `response.updateCompositeStatus(properties)` - merges the desired composite resource status in the
     function response.
     ```javascript
     export default function (req, rsp) {
       // ...skip for brevity
       rsp.updateCompositeStatus({ userCount: 1, message: 'All good' })
     }
     ```

## External dependencies

Because the function isn't based on Node.js or any other of the full-fledged JavaScript runtimes, it
doesn't support external dependencies or Node.js modules. However, users can use [ESBuild][esbuild],
or [Webpack][webpack], or any other similar tool to bundle external dependencies into a single JavaScript
file, and inject it into the composition pipeline as a single blob.

See [`external-dependencies`](examples/external-dependencies) example in the [`examples/`](./examples) folder.

For convenience, the runtime includes some "faux" external packages:

* `console` - implements some of the JavaScript's Console API static methods. The output is logged in the
  function container logs:
  ```javascript
  console.log('Hello');

  export default function (req, resp) {
    console.debug('Request', JSON.stringify(req));
    console.info('Info');
    console.warn('Warning');
    console.error('Error');
  }
  ```
* `btoa`, `atob` - functions for working with Base64 encoding:
  ```javascript
  const enc = btoa('string');
  const dec = atob(enc); // => 'string'
  ```

  **NB!** Unlike functions [`Window.btoa()`][base64] and [`Window.atob()`][base64] available
  in browsers, these functions work natively with UTF-8 strings and don't require additional
  manipulations:
  ```javascript
  // this will work in your composition function, but won't work in browsers
  btoa("a Ā 𐀀 文 🦄")
  ```
  
## Code transpilation

[Goja][goja] natively only supports ECMAScript 5.1 syntax, so in order to use modern syntax features,
the source code must be _transpiled_ into a ES 5.1 syntax. For convenience, transpilation is built-in
into the function server and is enabled by default.

For large functions, however, this additional pre-processing can impact performance, so if the function 
is already written in ES 5.1 compatible syntax (or pre-processed before injecting the source into a Composition),
you can disable server-side transpilation:

```yaml
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: function-javascript
spec:
  compositeTypeRef:
    apiVersion: example.crossplane.io/v1
    kind: XR
  mode: Pipeline
  pipeline:
  - step: run-the-template
    functionRef:
      name: function-javascript
    input:
      apiVersion: javascript.fn.crossplane.io/v1beta1
      kind: Input
      spec:
        source:
          transpile: false # <-- disable transpilation
          inline: |
            // source code
```

## Developing this function

This function uses [Go][go], [Docker][docker], and the [Crossplane CLI][cli] to
build functions.

```shell
# Run code generation - see input/generate.go
$ make generate

# Run tests - see fn_test.go
$ make test

# Build the function's runtime image - see Dockerfile
$ make img.build

# Build a function package - see package/crossplane.yaml
$ make xpkg.build
```

[functions]: https://docs.crossplane.io/latest/concepts/composition-functions
[go]: https://go.dev
[function guide]: https://docs.crossplane.io/knowledge-base/guides/write-a-composition-function-in-go
[package docs]: https://pkg.go.dev/github.com/crossplane/function-sdk-go
[docker]: https://www.docker.com
[cli]: https://docs.crossplane.io/latest/cli
[goja]: https://github.com/dop251/goja
[req]: https://buf.build/crossplane/crossplane/docs/main:apiextensions.fn.proto.v1beta1#apiextensions.fn.proto.v1beta1.RunFunctionRequest
[resp]: https://buf.build/crossplane/crossplane/docs/main:apiextensions.fn.proto.v1beta1#apiextensions.fn.proto.v1beta1.RunFunctionResponse
[esbuild]: https://esbuild.github.io/
[webpack]: https://webpack.js.org/
[base64]: https://developer.mozilla.org/en-US/docs/Glossary/Base64
[Babel]: https://babeljs.io/
