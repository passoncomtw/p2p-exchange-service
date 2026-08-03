#!/usr/bin/env python3
"""
make swagger 後執行，將 example 值 patch 進 goctl 產生的 spec.json。
goctl-swagger 不支援 example tag，所以用此腳本補上。
"""

import json
import sys

EXAMPLES = {
    "LoginRequest": {
        "username": "testdemo001",
        "password": "a12345678",
    },
    "LoginResponse": {
        "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOjEsInVzZXJuYW1lIjoidGVzdGRlbW8wMDEifQ.abc123",
        "expireIn": 86400,
    },
    "LoginUser": {
        "id": 1,
        "account": "testdemo001",
        "name": "Test User",
    },
    "VersionResponse": {
        "version": "v1.0.0",
    },
}

def patch(path: str) -> None:
    with open(path) as f:
        spec = json.load(f)

    definitions = spec.get("definitions", {})

    for def_name, examples in EXAMPLES.items():
        if def_name not in definitions:
            continue
        props = definitions[def_name].get("properties", {})
        # 修正 required：移除 goctl 誤塞的 default 值
        required = definitions[def_name].get("required", [])
        cleaned_required = [r for r in required if r in props]
        definitions[def_name]["required"] = cleaned_required
        # 補上 example
        for field, value in examples.items():
            if field in props:
                props[field]["example"] = value

    with open(path, "w") as f:
        json.dump(spec, f, indent=2, ensure_ascii=False)
        f.write("\n")

    print(f"patch-examples: {path} patched.")

if __name__ == "__main__":
    patch(sys.argv[1] if len(sys.argv) > 1 else "internal/swagger/dist/spec.json")
