#!/usr/bin/env python3
"""Extract Craft's embedded OpenAPI data without executing remote JavaScript."""
import argparse, hashlib, json, pathlib, re, urllib.request

class DataParser:
    def __init__(self, text): self.text, self.pos, self.refs = text, 0, {}
    def value(self):
        s = self.text
        while s[self.pos].isspace(): self.pos += 1
        ref = re.match(r'\$R\[(\d+)\]', s[self.pos:])
        if ref:
            self.pos += len(ref[0]); key = int(ref[1])
            if s[self.pos:self.pos+1] == '=':
                self.pos += 1; v = self.value(); self.refs[key] = v; return v
            if key not in self.refs: raise ValueError('Unresolved loader reference')
            return self.refs[key]
        c = s[self.pos]
        if c == '"':
            end = self.pos + 1
            while end < len(s):
                if s[end] == '\\': end += 2; continue
                if s[end] == '"': break
                end += 1
            token = s[self.pos:end+1]
            token = re.sub(r'\\x([0-9A-Fa-f]{2})', lambda m: '\\u00'+m[1], token)
            token = token.replace("\\'", "'").replace('\\v', '\\u000b')
            v = json.loads(token); self.pos = end + 1; return v
        if c in '{[':
            obj = {} if c == '{' else []; end = '}' if c == '{' else ']'; self.pos += 1
            while s[self.pos] != end:
                if c == '{':
                    if s[self.pos] == '"': key = self.value()
                    else:
                        match = re.match(r'(?:[A-Za-z_$][\w$]*|[0-9]+)', s[self.pos:])
                        if not match: raise ValueError('Invalid object key')
                        key = match[0]; self.pos += len(key)
                    if s[self.pos] != ':': raise ValueError('Expected colon')
                    self.pos += 1; obj[key] = self.value()
                else: obj.append(self.value())
                if s[self.pos] == ',': self.pos += 1
                elif s[self.pos] != end: raise ValueError('Invalid separator')
            self.pos += 1; return obj
        for token, value in [('!0', True), ('!1', False), ('true', True), ('false', False), ('null', None), ('void 0', None), ('undefined', None)]:
            if s.startswith(token, self.pos): self.pos += len(token); return value
        match = re.match(r'-?\d+(?:\.\d+)?(?:[eE][+-]?\d+)?', s[self.pos:])
        if not match: raise ValueError('Unsupported loader data near position '+str(self.pos))
        self.pos += len(match[0]); return json.loads(match[0])

def extract(html):
    start = html.find('{openapi:')
    if start < 0: raise ValueError('Craft page has no embedded OpenAPI document')
    result = DataParser(html[start:]).value()
    if not isinstance(result.get('paths'), dict): raise ValueError('Missing paths')
    return result

def canonical(spec): return json.dumps(spec, sort_keys=True, ensure_ascii=True, separators=(',', ':'))

def main():
    p = argparse.ArgumentParser(); p.add_argument('--url', default='https://connect.craft.do/link/HHRuPxZZTJ6/docs/v1'); p.add_argument('--html'); p.add_argument('--check', action='store_true'); p.add_argument('--output', default='docs/contracts/craft-rest-space-openapi.json'); args=p.parse_args()
    html = pathlib.Path(args.html).read_text() if args.html else urllib.request.urlopen(urllib.request.Request(args.url, headers={'User-Agent':'craft-cli-contract-check'}), timeout=45).read().decode()
    spec = extract(html); target=pathlib.Path(args.output)
    digest=hashlib.sha256(canonical(spec).encode()).hexdigest()
    if args.check:
        old=json.loads(target.read_text())
        if canonical(old)!=canonical(spec): raise SystemExit('Craft contract changed. Review the live schema and update fixtures before refreshing the pin.')
    else: target.write_text(json.dumps(spec,indent=2,ensure_ascii=True)+'\n')
    print('Craft contract SHA256: '+digest)
if __name__=='__main__':main()
