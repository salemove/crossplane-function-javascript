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
            import * as Base64 from 'base64';
             
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
                // expose some connection details, get value from a resource generated within this function
                rsp.setConnectionDetails({ bucketName: req.observed.resources.bucket.resource.metadata.name });
             
                // patch composite resource status
                rsp.updateCompositeStatus({ bucketName: req.observed.resources.bucket.resource.metadata.name });
              }
            };
  - step: automatically-detect-ready-composed-resources
    functionRef:
       name: function-auto-ready
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

     Note that connection details from other observed resources are base64-encoded, and
     if you want to use them in other composed resources, or in the composition connection
     details, you need to decode them first:
     ```javascript
     import * as Base64 from 'base64';

     export default function (req, rsp) {
       // ...skip for brevity
       const username = Base64.decode(req.observed.resources.user.connectionDetails.username);
       rsp.setConnectionDetails({ username });
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
* `base64` - includes functions for working with Base64 encoding:
  ```javascript
  import * as Base64 from 'base64';
  const enc = Base64.encode('string');
  const dec = Base64.decode(enc); // => 'string'
  ```
* `yaml` - includes functions for encoding and decoding objects into YAML format:
  ```javascript
  import * as YAML from 'yaml';
  const enc = YAML.stringify(someObject);
  const dec = YAML.parse(enc);
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
