// Code generated by "gosl"; DO NOT EDIT
// kernel: Compute

@group(0) @binding(0)
var<storage, read_write> Seed: array<Seeds>;
@group(0) @binding(1)
var<storage, read_write> Data: array<Rnds>;

@compute @workgroup_size(64, 1, 1)
fn main(@builtin(global_invocation_id) idx: vec3<u32>) {
	Compute(idx.x);
}


///////////// import: "rand.go"
struct Seeds {
	Seed: su64,
	pad:  i32,
	pad1: i32,
}
struct Rnds {
	Uints: vec2<u32>,
	pad:   i32,
	pad1: i32,
	Floats: vec2<f32>,
	pad2:   i32,
	pad3: i32,
	Floats11: vec2<f32>,
	pad4:     i32,
	pad5: i32,
	Gauss: vec2<f32>,
	pad6:  i32,
	pad7: i32,
}
fn Rnds_RndGen(r: ptr<function,Rnds>, counter: su64, idx: u32) {
	(*r).Uints = RandUint32Vec2(counter, u32(0), idx);
	(*r).Floats = RandFloat32Vec2(counter, u32(1), idx);
	(*r).Floats11 = RandFloat32Range11Vec2(counter, u32(2), idx);
	(*r).Gauss = RandFloat32NormVec2(counter, u32(3), idx);
}
fn Compute(i: u32) { //gosl:kernel
	var data=Data[i];
	Rnds_RndGen(&data, Seed[0].Seed, i);
	Data[i]=data;
}

///////////// import: "slrand.wgsl"
fn Philox2x32round(counter: su64, key: u32) -> su64 {
	let mul = Uint32Mul64(u32(0xD256D193), counter.x);
	var ctr: su64;
	ctr.x = mul.y ^ key ^ counter.y;
	ctr.y = mul.x;
	return ctr;
}
fn Philox2x32bumpkey(key: u32) -> u32 {
	return key + u32(0x9E3779B9);
}
fn Philox2x32(counter: su64, key: u32) -> vec2<u32> {
	var ctr = Philox2x32round(counter, key); // 1
	var ky = Philox2x32bumpkey(key);
	ctr = Philox2x32round(ctr, ky); // 2
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 3
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 4
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 5
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 6
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 7
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 8
	ky = Philox2x32bumpkey(ky);
	ctr = Philox2x32round(ctr, ky); // 9
	ky = Philox2x32bumpkey(ky);
	return Philox2x32round(ctr, ky); // 10
}
fn RandUint32Vec2(counter: su64, funcIndex: u32, key: u32) -> vec2<u32> {
	return Philox2x32(Uint64Add32(counter, funcIndex), key);
}
fn RandUint32(counter: su64, funcIndex: u32, key: u32) -> u32 {
	return Philox2x32(Uint64Add32(counter, funcIndex), key).x;
}
fn RandFloat32Vec2(counter: su64, funcIndex: u32, key: u32) -> vec2<f32> {
	return Uint32ToFloat32Vec2(RandUint32Vec2(counter, funcIndex, key));
}
fn RandFloat32(counter: su64, funcIndex: u32, key: u32) -> f32 { 
	return Uint32ToFloat32(RandUint32(counter, funcIndex, key));
}
fn RandFloat32Range11Vec2(counter: su64, funcIndex: u32, key: u32) -> vec2<f32> {
	return Uint32ToFloat32Vec2(RandUint32Vec2(counter, funcIndex, key));
}
fn RandFloat32Range11(counter: su64, funcIndex: u32, key: u32) -> f32 { 
	return Uint32ToFloat32Range11(RandUint32(counter, funcIndex, key));
}
fn RandBoolP(counter: su64, funcIndex: u32, key: u32, p: f32) -> bool { 
	return (RandFloat32(counter, funcIndex, key) < p);
}
fn sincospi(x: f32) -> vec2<f32> {
	let PIf = 3.1415926535897932;
	var r: vec2<f32>;
	r.x = cos(PIf*x);
	r.y = sin(PIf*x);
	return r;
}
fn RandFloat32NormVec2(counter: su64, funcIndex: u32, key: u32) -> vec2<f32> { 
	let ur = RandUint32Vec2(counter, funcIndex, key);
	var f = sincospi(Uint32ToFloat32Range11(ur.x));
	let r = sqrt(-2.0 * log(Uint32ToFloat32(ur.y))); // guaranteed to avoid 0.
	return f * r;
}
fn RandFloat32Norm(counter: su64, funcIndex: u32, key: u32) -> f32 { 
	return RandFloat32Vec2(counter, funcIndex, key).x;
}
fn RandUint32N(counter: su64, funcIndex: u32, key: u32, n: u32) -> u32 { 
	let v = RandFloat32(counter, funcIndex, key);
	return u32(v * f32(n));
}
struct RandCounter {
	Counter: su64,
	HiSeed: u32,
	pad: u32,
}
fn RandCounter_Reset(ct: ptr<function,RandCounter>) {
	(*ct).Counter.x = u32(0);
	(*ct).Counter.y = (*ct).HiSeed;
}
fn RandCounter_Seed(ct: ptr<function,RandCounter>, seed: u32) {
	(*ct).HiSeed = seed;
	RandCounter_Reset(ct);
}
fn RandCounter_Add(ct: ptr<function,RandCounter>, inc: u32) {
	(*ct).Counter = Uint64Add32((*ct).Counter, inc);
}

///////////// import: "sltype.wgsl"
alias su64 = vec2<u32>;
fn Uint32Mul64(a: u32, b: u32) -> su64 {
	let LOMASK = (((u32(1))<<16)-1);
	var r: su64;
	r.x = a * b;               /* full low multiply */
	let ahi = a >> 16;
	let alo = a & LOMASK;
	let bhi = b >> 16;
	let blo = b & LOMASK;
	let ahbl = ahi * blo;
	let albh = alo * bhi;
	let ahbl_albh = ((ahbl&LOMASK) + (albh&LOMASK));
	var hit = ahi*bhi + (ahbl>>16) +  (albh>>16);
	hit += ahbl_albh >> 16; /* carry from the sum of lo(ahbl) + lo(albh) ) */
	/* carry from the sum with alo*blo */
	if ((r.x >> u32(16)) < (ahbl_albh&LOMASK)) {
		hit += u32(1);
	}
	r.y = hit; 
	return r;
}
/*
fn Uint32Mul64(a: u32, b: u32) -> su64 {
	return su64(a) * su64(b);
}
*/
fn Uint64Add32(a: su64, b: u32) -> su64 {
	if (b == 0) {
		return a;
	}
	var s = a;
	if (s.x > u32(0xffffffff) - b) {
		s.y++;
		s.x = (b - 1) - (u32(0xffffffff) - s.x);
	} else {
		s.x += b;
	}
	return s;
}
fn Uint64Incr(a: su64) -> su64 {
	var s = a;
	if(s.x == 0xffffffff) {
		s.y++;
		s.x = u32(0);
	} else {
		s.x++;
	}
	return s;
}
fn Uint32ToFloat32(val: u32) -> f32 {
	let factor = f32(1.0) / (f32(u32(0xffffffff)) + f32(1.0));
	let halffactor = f32(0.5) * factor;
	var f = f32(val) * factor + halffactor;
	if (f == 1.0) { // exclude 1
		return bitcast<f32>(0x3F7FFFFF);
	}
	return f;
}
fn Uint32ToFloat32Vec2(val: vec2<u32>) -> vec2<f32> {
	var r: vec2<f32>;
	r.x = Uint32ToFloat32(val.x);
	r.y = Uint32ToFloat32(val.y);
	return r;
}
fn Uint32ToFloat32Range11(val: u32) -> f32 {
	let factor = f32(1.0) / (f32(i32(0x7fffffff)) + f32(1.0));
	let halffactor = f32(0.5) * factor;
	return (f32(val) * factor + halffactor);
}
fn Uint32ToFloat32Range11Vec2(val: vec2<u32>) -> vec2<f32> {
	var r: vec2<f32>;
	r.x = Uint32ToFloat32Range11(val.x);
	r.y = Uint32ToFloat32Range11(val.y);
	return r;
}