#!/bin/bash

# Test the file parser with a sample Layer program

cat > /tmp/test.layer << 'EOF'
@layers [core, debug]

@layer core {
    fn factorial(n) =
        match n with
        | 0 -> 1
        | n -> n * factorial(n - 1)
}

@layer debug {
    fn factorial(n) =
        let _ = log("Computing factorial of {n}") in
            core.factorial(n)
}

factorial(5)
EOF

echo "Test .layer file created:"
cat /tmp/test.layer

echo ""
echo "You can test the parseLayerFile function in the browser console:"
echo "const result = window.parseLayerFile(\`$(cat /tmp/test.layer)\`)"
echo "console.log(result)"
