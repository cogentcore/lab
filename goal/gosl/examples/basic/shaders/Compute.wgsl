// Code generated by "gosl"; DO NOT EDIT
// kernel: Compute

// // Params are the parameters for the computation. 
@group(0) @binding(0)
var<storage, read_write> Params: array<ParamStruct>;
@group(0) @binding(1)
var<storage, read_write> Data: array<DataStruct>;

@compute @workgroup_size(64, 1, 1)
fn main(@builtin(global_invocation_id) idx: vec3<u32>) {
	Compute(idx.x);
}


///////////// import: "compute.wgsl"

//gosl:import "cogentcore.org/core/math32"

//gosl:vars

// Params are the parameters for the computation.
//gosl:read-only

// Data is the data on which the computation operates.

// DataStruct has the test data
struct DataStruct {

	// raw value
	Raw: f32,

	// integrated value
	Integ: f32,

	// exp of integ
	Exp: f32,

	// must pad to multiple of 4 floats for arrays
	pad: f32,
}

// ParamStruct has the test params
struct ParamStruct {

	// rate constant in msec
	Tau: f32,

	// 1/Tau
	Dt: f32,

	pad:  f32,
	pad1: f32,
}

// IntegFromRaw computes integrated value from current raw value
fn ParamStruct_IntegFromRaw(ps: ptr<function,ParamStruct>, ds: ptr<function,DataStruct>) {
	(*ds).Integ += (*ps).Dt * ((*ds).Raw - (*ds).Integ);
	(*ds).Exp = FastExp(-(*ds).Integ);
}

// Compute does the main computation
fn Compute(i: u32) { //gosl:kernel
	// Params[0].IntegFromRaw(&Data[i])
	var params = Params[0];
	var data = Data[i];
	ParamStruct_IntegFromRaw(&params, &data);
	Data[i] = data;
}


///////////// import: "fastexp.wgsl"

// FastExp is a quartic spline approximation to the Exp function, by N.N. Schraudolph
// It does not have any of the sanity checking of a standard method -- returns
// nonsense when arg is out of range.  Runs in 2.23ns vs. 6.3ns for 64bit which is faster
// than Exp actually.
fn FastExp(x: f32) -> f32 {
	if (x <= -88.02969) { // this doesn't add anything and -exp is main use-case anyway
		return f32(0.0);
	}
	var i = i32(12102203*x) + i32(127)*(i32(1)<<23);
	var m = i >> 7 & 0xFFFF; // copy mantissa
	i += (((((((((((3537 * m) >> 16) + 13668) * m) >> 18) + 15817) * m) >> 14) - 80470) * m) >> 11);
	return bitcast<f32>(u32(i));
}
