import YAML from 'yaml';

export default (req, rsp) => {
  console.log("Running function");

  ["us-east-1", "us-east-2"].forEach((region) => {
    rsp.setDesiredComposedResource(`bucket-${region}`, {
      apiVersion: 'example.org/v1alpha1',
      kind: 'Bucket',
      metadata: {
        annotations: {
          'javascript.fn.crossplane.io/ready': 'True',
          ...req.input.metadata.annotations
        },
        labels: {
          ...(req.observed.composite.resource.metadata?.labels || {}),
          foo: "bar"
        },
        spec: {
          name: `test-${region}`,
          region: region,
          compositeRegion: req.observed.composite.resource.spec.region,
          yamlTest: YAML.stringify({ foo: "bar", bax: [12, 23] }),
          b64test: btoa('abcdefgh')
        }
      }
    });
  });

  rsp.setConnectionDetails({ "foo": "bar" })
  rsp.updateCompositeStatus({
    something: "in the way",
    she: { moves: true }
  });

  rsp.updateCompositeStatus({ x: "y" });
};
