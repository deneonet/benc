const MaxDepth = 2;

function randIntn(n) {
    return Math.floor(Math.random() * n);
}

function randInt() {
    return Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);
}

function randInt31() {
    return Math.floor(Math.random() * 2147483647);
}

function randUint32() {
    return Math.floor(Math.random() * 4294967296);
}

function RandomCount() {
    return 1 + randIntn(3);
}

function GenerateBool(_depth) {
    return randIntn(2) === 1;
}

function GenerateInt(_depth) {
    const v = randInt();
    return randIntn(2) === 0 ? v : -v;
}

function GenerateInt8(_depth) {
    return (randIntn(256) - 128);
}

function GenerateInt16(_depth) {
    return (randIntn(65536) - 32768);
}

function GenerateInt32(_depth) {
    const v = randInt31();
    return randIntn(2) === 0 ? v : -v;
}

function GenerateInt64(_depth) {
    const low = BigInt(randUint32());
    const high = BigInt(randUint32());
    const val = (high << 32n) | low;
    return randIntn(2) === 0 ? val : -val;
}

function GenerateUint(_depth) {
    return randInt();
}

function GenerateUint8(_depth) {
    return randIntn(256);
}

function GenerateUint16(_depth) {
    return randIntn(65536);
}

function GenerateUint32(_depth) {
    return randUint32();
}

function GenerateUint64(_depth) {
    const low = BigInt(randUint32());
    const high = BigInt(randUint32());
    return (high << 32n) | low;
}

function GenerateUintptr(_depth) {
    return GenerateUint64(_depth);
}

function GenerateFloat32(_depth) {
    return Math.fround(Math.random() * 1000); 
}

function GenerateFloat64(_depth) {
    return Math.random() * 1000;
}

function GenerateComplex64(_depth) {
    return { r: GenerateFloat32(), i: GenerateFloat32() };
}

function GenerateComplex128(_depth) {
    return { r: GenerateFloat64(), i: GenerateFloat64() };
}

function GenerateString(_depth) {
    return RandomString(5 + randIntn(15));
}

function GenerateByte(_depth) {
    return randIntn(256);
}

function GenerateBytes(_depth) {
    return RandomBytes(3 + randIntn(7));
}

function GenerateRune(_depth) {
    return randInt31();
}

function GenerateTime(_depth) {
    return RandomTime();
}

function GenerateSlice(depth, generator) {
    const count = RandomCount();
    const s = new Array(count);
    for (let i = 0; i < count; i++) {
        s[i] = generator(depth);
    }
    return s;
}

function GenerateSliceSlice(depth, generator) {
    const count = RandomCount();
    const s = new Array(count);
    for (let i = 0; i < count; i++) {
        s[i] = GenerateSlice(depth, generator);
    }
    return s;
}

function BytesEqual(a, b) {
    if (a.length !== b.length) {
        return false;
    }
    for (let i = 0; i < a.length; i++) {
        if (a[i] !== b[i]) {
            return false;
        }
    }
    return true;
}

function GeneratePointer(depth, generator) {
    if (randIntn(4) === 0) {
        return null;
    }
    return generator(depth);
}

function GenerateMap(depth, keyGen, valGen) {
    const count = RandomCount();
    const m = new Map();
    for (let i = 0; i < count; i++) {
        m.set(keyGen(depth), valGen(depth));
    }
    return m;
}

function RandomString(length) {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789";
    let res = "";
    for (let i = 0; i < length; i++) {
        res += charset[randIntn(charset.length)];
    }
    return res;
}

function RandomBytes(length) {
    const b = new Uint8Array(length);
    for (let i = 0; i < length; i++) {
        b[i] = randIntn(256);
    }
    return b;
}

function RandomTime() {
    const now = Date.now();
    const offset = randIntn(1000000) * 1000; // milliseconds
    return new Date(now + offset);
}

function RandomTimePtr() {
    return RandomTime();
}

function CompareField(name, cmp) {
    const err = cmp();
    if (err) {
        return new Error(`${name}: ${err.message}`);
    }
    return null;
}

function ComparePrimitive(a, b) {
    if (a !== b) {
        return new Error(`mismatch: ${a} != ${b}`);
    }
    return null;
}

function CompareBytes(a, b) {
    if (!BytesEqual(a, b)) {
        const hexA = Array.from(a).map(byte => byte.toString(16).padStart(2, '0')).join('');
        const hexB = Array.from(b).map(byte => byte.toString(16).padStart(2, '0')).join('');
        return new Error(`mismatch: ${hexA} != ${hexB}`);
    }
    return null;
}

function CompareSlice(a, b, elemCmp) {
    if (a.length !== b.length) {
        return new Error(`length mismatch: ${a.length} != ${b.length}`);
    }
    if (a.length === 0 && b.length === 0) {
        return null;
    }
    for (let i = 0; i < a.length; i++) {
        const err = elemCmp(a[i], b[i]);
        if (err) {
            return new Error(`index [${i}]: ${err.message}`);
        }
    }
    return null;
}

function CompareMap(a, b, valCmp) {
    if (a.size !== b.size) {
        return new Error(`length mismatch: ${a.size} != ${b.size}`);
    }
    if (a.size === 0 && b.size === 0) {
        return null;
    }
    for (const [k, v1] of a) {
        if (!b.has(k)) {
            return new Error(`key ${k} missing in b`);
        }
        const v2 = b.get(k);
        const err = valCmp(v1, v2);
        if (err) {
            return new Error(`key ${k} value: ${err.message}`);
        }
    }
    return null;
}

function ComparePointer(a, b, elemCmp) {
    if (a === null && b === null) {
        return null;
    }
    if (a === null && b !== null) {
        return new Error("mismatch: a is nil, b is not");
    }
    if (a !== null && b === null) {
        return new Error("mismatch: a is not nil, b is nil");
    }
    return elemCmp(a, b);
}

if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        MaxDepth,
        RandomCount,
        GenerateBool,
        GenerateInt, GenerateInt8, GenerateInt16, GenerateInt32, GenerateInt64,
        GenerateUint, GenerateUint8, GenerateUint16, GenerateUint32, GenerateUint64, GenerateUintptr,
        GenerateFloat32, GenerateFloat64,
        GenerateComplex64, GenerateComplex128,
        GenerateString,
        GenerateByte, GenerateBytes,
        GenerateRune,
        GenerateTime,
        GenerateSlice, GenerateSliceSlice,
        GeneratePointer,
        GenerateMap,
        RandomString, RandomBytes, RandomTime, RandomTimePtr,
        BytesEqual,
        CompareField, ComparePrimitive, CompareBytes, CompareSlice, CompareMap, ComparePointer
    };
}