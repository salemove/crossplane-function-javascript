import process from 'node:process';
import fs from 'node:fs/promises';
import * as esbuild from 'esbuild';
import babel from '@babel/core';
import YAML from 'yaml';

const buildResult = await esbuild.build({
  entryPoints: ['src/index.js'],
  bundle: true,
  write: false,
  platform: "neutral",
  sourcemap: "inline",
  sourceRoot: "examples/dependencies",
  target: "es6"
});

if (buildResult.errors.length > 0) {
  console.log(buildResult.errors);
  process.exit(1);
}

const transpiled = await babel.transformAsync(buildResult.outputFiles[0].text, {
  plugins: [
    ["@babel/plugin-transform-modules-commonjs", { loose: true }],
  ],
  ast: false,
  babelrc: false,
  sourceMaps: 'inline',
  inputSourceMap: true,
  compact: false,
  retainLines: true,
  highlightCode: false
});

const composition = {
  apiVersion: 'apiextensions.crossplane.io/v1',
  kind: 'Composition',
  metadata: { name: 'function-javascript' },
  spec: {
    compositeTypeRef: {
      apiVersion: 'example.crossplane.io/v1',
      kind: 'XR'
    },
    mode: 'Pipeline',
    pipeline: [{
      step: 'run-the-template',
      functionRef: {
        name: 'function-javascript',
      },
      input: {
        apiVersion: 'javascript.fn.crossplane.io/v1beta1',
        kind: 'Input',
        metadata: {
          annotations: { key: 'value' },
        },
        spec: {
          source: {
            transpile: false,
            inline: transpiled.code
          }
        }
      }
    }]
  }
};

await fs.writeFile('./composition.yaml', YAML.stringify(composition));
