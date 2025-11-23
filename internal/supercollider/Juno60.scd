(
var mainCaller,
noteArrayJuno60 = Array.newClear(128);

mainCaller = { arg set, k, v, synBefore, note, vel, pan,
	attack, release, duration, trackVolume, filter,
	effectDryOut, effectCombOut, effectReverbOut, effectReverb, effectComb, trackOut,
	// Juno-60 specific parameters
	amp = 1.0,
	decay = 0.5,
	sustain = 0.5,
	cutoff = 0.5,
	resonance = 0.0,
	envMod = 0.0,
	saw = 1,
	pulse = 0.0,
	sub = 0.0,
	noise = 0.0,
	pwm = 0.5,
	chorus = 0,
	chorusDrywet = 1.0,
	lfoRate = 0.5,
	lfoDelay = 0.0,
	lfoAuto = 0,
	dcoLfo = 0.0,
	dcoPwmMod = 0,
	dcoRange = 1.0,
	hpf = 0.0,
	vcfLfo = 0.0,
	vcfKey = 1.0,
	vcfDir = 1.0,
	vcaType = 0.0,
	vcaValue = 0.5;

	// set 0 = note on/off, 1 = parameter change, 2 = all notes off
	if (set>0, {
		if (set>1, {
			// All notes off
			128.do({ arg noteSet;
				if (noteArrayJuno60[noteSet].notNil, {
					if (noteArrayJuno60[noteSet].isRunning, {
						noteArrayJuno60[noteSet].set(\gate, 0);
					});
				});
			});
		}, {
			// Parameter change for running notes
			128.do({ arg noteSet;
				if (noteArrayJuno60[noteSet].notNil, {
					if (noteArrayJuno60[noteSet].isRunning, {
						noteArrayJuno60[noteSet].set(k.asString, v);
					});
				});
			});
		});
	}, {
		// Note on/off
		if (vel > 0, {
			// Note on
			if (noteArrayJuno60[note] !== nil, {
				noteArrayJuno60[note].set(\gate, 0);
			});

			noteArrayJuno60[note] = Synth(\Juno60, [
				\gate, 1,
				\note, note,
				\amp, amp,
				\attack, attack,
				\decay, decay,
				\sustain, sustain,
				\release, release,
				\cutoff, cutoff,
				\resonance, resonance,
				\envMod, envMod,
				\saw, saw,
				\pulse, pulse,
				\sub, sub,
				\noise, noise,
				\pwm, pwm,
				\chorus, chorus,
				\chorusDrywet, chorusDrywet,
				\lfoRate, lfoRate,
				\lfoDelay, lfoDelay,
				\lfoAuto, lfoAuto,
				\dcoLfo, dcoLfo,
				\dcoPwmMod, dcoPwmMod,
				\dcoRange, dcoRange,
				\hpf, hpf,
				\vcfLfo, vcfLfo,
				\vcfKey, vcfKey,
				\vcfDir, vcfDir,
				\vcaType, vcaType,
				\vcaValue, vcaValue,
				\pan, pan,
				\duration, duration,
				\trackVolume, trackVolume,
				\effectDryOut, effectDryOut,
				\effectCombOut, effectCombOut,
				\effectReverbOut, effectReverbOut,
				\effectReverb, effectReverb,
				\effectComb, effectComb,
				\trackOut, trackOut
			], synBefore, \addBefore);

			NodeWatcher.register(noteArrayJuno60[note]);
		}, {
			// Note off
			noteArrayJuno60[note].set(\gate, 0);
			noteArrayJuno60[note] = nil;
		});
	});
};

mainCaller;
)
