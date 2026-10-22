#!/usr/bin/env python3
"""Validate offline CLI schemas and realistic examples without executing mutations."""
import json, pathlib, shlex, subprocess, sys
import jsonschema
binary = str(pathlib.Path(sys.argv[1]).resolve())
manifest = json.loads(subprocess.check_output([binary, 'schema']))
for command in manifest['commands']:
    schema = {'allOf': [command['stdout_schema']], 'extensions': manifest['extensions']}
    jsonschema.Draft202012Validator.check_schema(schema)
    for example in command['extensions']['examples']:
        argv = shlex.split(example.strip())[1:]
        input_schema = json.loads(subprocess.check_output([binary, *argv, '--cli-request-schema']))
        jsonschema.Draft202012Validator.check_schema(input_schema)
        # Validate positional cardinality separately from flag parsing, which the binary performs.
        positional = []
        flags = input_schema['properties']['flags']['properties']
        index = len(command['name'].split())
        while index < len(argv):
            token = argv[index]
            if token.startswith('--'):
                name = token[2:].split('=')[0]
                if '=' not in token and flags[name].get('type') != 'boolean': index += 1
            elif token.startswith('-'):
                raise AssertionError('Use canonical long flags in generated examples: '+example)
            else: positional.append(token)
            index += 1
        jsonschema.validate(positional, input_schema['properties']['args'])
for name in ['version', 'limits', 'context', 'profiles list', 'skill-path', 'llm', 'audit agent-dx']:
    command = next(c for c in manifest['commands'] if c['name'] == name)
    output = json.loads(subprocess.check_output([binary, *name.split()]))
    jsonschema.validate(output, {'allOf': [command['stdout_schema']], 'extensions': manifest['extensions']})
print('CLI contracts, examples and local output schemas passed for', len(manifest['commands']), 'commands')
