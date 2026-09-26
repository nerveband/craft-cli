#!/usr/bin/env python3
"""Validate the generated manifest against the pinned frozen CLI Spec schema."""
import json
from pathlib import Path
import jsonschema
root = Path(__file__).resolve().parents[1]
jsonschema.validate(json.loads((root / "docs/command-reference.json").read_text()), json.loads((root / "docs/contracts/clispec-v0.2.schema.json").read_text()))
print("CLI Spec v0.2 schema validation passed; behavioral conformance is audited separately")
