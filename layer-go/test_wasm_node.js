// Test WASM parseLayerFile function with Node.js
const fs = require('fs');

// Load the Go WASM runtime
require('../visual-editor/frontend/public/wasm_exec.js');

async function testParseLayerFile() {
    const go = new Go();

    // Load the WASM module
    const wasmPath = '../visual-editor/frontend/public/layer.wasm';
    const wasmBuffer = fs.readFileSync(wasmPath);
    const result = await WebAssembly.instantiate(wasmBuffer, go.importObject);

    // Run the Go WASM module (this sets up the exports)
    go.run(result.instance);

    // Wait a bit for initialization
    await new Promise(resolve => setTimeout(resolve, 100));

    // Test the parseLayerFile function
    const testSource = `@layers [core, debug]

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

factorial(5)`;

    console.log('Testing parseLayerFile...\n');

    if (typeof global.parseLayerFile !== 'function') {
        console.error('ERROR: parseLayerFile is not available!');
        console.log('Available globals:', Object.keys(global).filter(k => k.startsWith('parse') || k.startsWith('execute')));
        process.exit(1);
    }

    const parsed = global.parseLayerFile(testSource);

    console.log('Parse result:', JSON.stringify(parsed, null, 2));

    // Verify the results
    if (parsed.success) {
        console.log('\n✓ Success: true');
        console.log('✓ Enabled layers:', parsed.enabledLayers);
        console.log('✓ Number of layers:', Object.keys(parsed.layers).length);
        console.log('\n--- Layer Code ---');
        for (const [layerName, code] of Object.entries(parsed.layers)) {
            console.log(`\nLayer "${layerName}":`);
            console.log(code);
        }
        console.log('\n--- Main Expression ---');
        console.log(parsed.mainExpression);

        // Check if layer code is actually extracted (not empty)
        const coreCode = parsed.layers.core || '';
        if (coreCode.includes('factorial')) {
            console.log('\n✓ Layer code correctly extracted!');
        } else {
            console.log('\n✗ ERROR: Layer code is empty or incorrect');
            process.exit(1);
        }
    } else {
        console.log('\n✗ Parse failed:', parsed.error);
        process.exit(1);
    }
}

testParseLayerFile().catch(err => {
    console.error('Test failed:', err);
    process.exit(1);
});
