(
s.waitForBoot({
Routine{
~serverLatency = 0.1;
~synthPlayback = nil;
~synthRecord = Dictionary.new();
~samplesPlaying = Dictionary.new();
~synthsPlaying = Dictionary.new();

    	SynthDef("SuperSaw",{
    		arg vibrRate = 6, vibrDepth = 0.3, drive = 1.5, detune = 0.2, spread = 0.6, lpenv = 0, lpa = 0;
    		var ducked;
    		var cutoff = \lowPassFilter.kr(20000);
    		var freq = \note.kr(60).min(127).max(0).midicps;
    		var env = EnvGen.ar(
    			Env.adsr(
    				\attack.kr(0.1),
    				\decay.kr(1.0),
    				\sustain.kr(0.7),
    				\release.kr(0.5)
    			),
    			1-EnvGen.kr(Env.new([0,0,1],[\duration.kr(1),0])),
    			doneAction:2,
    		);
    		var snd, sig, freqs, unison = 7;

    		// Unison voices with random phases
    		freqs = Array.fill(unison, { |i|
    			var det = ((i / (unison - 1)) - 0.5) * detune; // ±detune/2
    			(freq) * (1 + (LFTri.kr(vibrRate,Rand(0,1)).range(-1, 1) * (2 ** (vibrDepth / 12) - 1))) * (2 ** (det / 12));
    		});

    		sig = Mix.fill(unison, { |i|
    			var osc = SawDPW.ar(freqs[i], Rand(0, 1)); // random phase
    			Pan2.ar(osc, ((i % 2) * 2 - 1) * spread);
    		}) / sqrt(unison);

    		snd = MoogLadder.ar(sig * drive,
    			10 + (XLine.kr(cutoff,((2**lpenv)*cutoff).min(18000).max(cutoff),\duration.kr(1)*lpa)*env)
    		) * (1 / drive);

    		snd = snd * env * \trackVolume.kr(0).dbamp * (1.0 / \noteSize.kr(1).sqrt)
    			* \velocity.kr(100).min(127).max(0).linlin(0,127,-24,24).dbamp;
    		snd = Balance2.ar(snd[0],snd[1],\pan.kr(0));

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);

    		Out.ar(\trackOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectDryOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectCombOut.kr(0), snd*\effectComb.kr(0));
    		Out.ar(\effectReverbOut.kr(0), snd*\effectReverb.kr(0));
    	}).add;


    	SynthDef("Open303",{
    		arg t_trig=1;
	        var snd, env;
    		var ducked;
	        snd = Open303.ar(
	            gate: \gate.kr(1),
	            trig: t_trig,
	            noteNum: \note.kr(60),
	            velocity: (\velocity.kr(100)>64.0),
	            waveform: \waveform.kr(0.2), // OPTION: waveform 0.0-1.0, default 0.5, coarse 0.1, fine 0.01
	            tuning: \tuning.kr(440), // OPTION: tuning 110-880, default 440, coarse 10, fine 1
	            cutoff: \lowPassFilter.kr(1000),
	            resonance: \resonance.kr(85), // OPTION: resonance 1-200, default 50, coarse 10, fine 1
				envMod: \envMod.kr(60), // OPTION: 0–100, default 50, coarse 10, fine 1
				decay: \normalDecay.kr(800), // OPTION: 30–3000, default 1000, coarse 100, fine 10
				accent: \accent.kr(70), // OPTION: 0–100, default 50, coarse 10, fine 1
				pregain: \basePregain.kr(6.neg), // OPTION: -60–0, default -12, coarse 6, fine 1
				postgain: \basePostgain.kr(6.neg), // OPTION: -60–0, default -12, coarse 6, fine 1
				distortion: \baseDrive.kr(0), // OPTION: 0–100, default 0, coarse 10, fine 1
				ampSustain: \ampSustain.kr(0), // OPTION: -inf–0, default 0, coarse 6, fine 1
				ampDecay: \ampDecay.kr(1230), // OPTION: 30–3000, default 1230, coarse 100, fine 10
				ampRelease: \ampRelease.kr(1), // OPTION: 1–2000, default 1, coarse 100, fine 1
				feedbackHPF: \feedbackHPF.kr(150), // OPTION: 20–2000, default 150, coarse 100, fine 10
				normalAttack: \normalAttack.kr(3), // OPTION: 0.3–30, default 3, coarse 1, fine 0.1
				accentAttack: \accentAttack.kr(3), // OPTION: 0.3–30, default 3, coarse 1, fine 0.1
				accentDecay: \accentDecay.kr(200), // OPTION: 30–3000, default 200, coarse 100, fine 10
				slideTime: \slideTime.kr(60), // OPTION: 0–300, default 60, coarse 10, fine 1
	        );
			env = EnvGen.ar(
				Env.perc(\attack.kr(0.1), \duration.kr(1)), t_trig, doneAction:2
			);

    		snd = snd * env * \trackVolume.kr(0).dbamp * (1.0 / \noteSize.kr(1).sqrt);
    		snd = Pan2.ar(snd,\pan.kr(0));

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);

    		Out.ar(\trackOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectDryOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectCombOut.kr(0), snd*\effectComb.kr(0));
    		Out.ar(\effectReverbOut.kr(0), snd*\effectReverb.kr(0));
    	}).add;


    	SynthDef("TB303",{
    		arg t_trig=1;
    		var osc, filt, preDrive, postDrive;
			var snd;
    		var ducked;
    		var cutoff = \lowPassFilter.kr(20000);
    		var res = \resonance.kr(0.5).clip(0.1, 3.0);
    		var freq = \note.kr(60).min(127).max(0).midicps;
    		var glideFreq = Lag.kr(freq, \glide.kr(0.0).max(0.001));
    		var waveMix = Lag.kr(\mixWave.kr(0.5).clip(0,1));

    		// var env = EnvGen.ar(
    		// 	Env.adsr(
    		// 		\attack.kr(0.1),
    		// 		\decay.kr(1.0),
    		// 		\sustain.kr(0.7),
    		// 		\release.kr(0.5)
    		// 	),
    		// 	1-EnvGen.kr(Env.new([0,0,1],[\duration.kr(1),0])),
    		// 	doneAction:2,
    		// );
			var env = EnvGen.ar(
				Env.perc(\attack.kr(0.1), \duration.kr(1)), t_trig, doneAction:2
			);
    		// var env2 = EnvGen.ar(Env.perc(\attack.kr(0.1), \release.kr(0.5)), t_trig);
    		var fenv = EnvGen.kr(Env.perc(0.001,\envRelease.kr(0.2)*\duration.kr(1)), t_trig) * \envAmt.kr(1.0) * (1 + (\accent.kr(0) * 1.4));


    		osc = Mix([
    			DPW4Saw.ar(glideFreq),
    			Pulse.ar(glideFreq, 0.5)
    		] * [1 - waveMix, waveMix]);

    		preDrive = tanh(osc * (2.5 + \accent.kr(0)));

    		filt = DFM1.ar(
    			preDrive,
    			(cutoff * (1 + fenv)).clip(80, 15000), // safe range
    			(res ).clip(0.2,3.0),
    			\drive.kr(1.0)
    		);

    		postDrive = HPF.ar(tanh(filt * (3 + \accent.kr(0))), 80);

    		postDrive = postDrive * env;

    		snd = postDrive * env * \trackVolume.kr(0).dbamp * (1.0 / \noteSize.kr(1).sqrt)
    			* \velocity.kr(100).min(127).max(0).linlin(0,127,-24,24).dbamp;
    		snd = Pan2.ar(snd,\pan.kr(0));

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);

    		Out.ar(\trackOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectDryOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectCombOut.kr(0), snd*\effectComb.kr(0));
    		Out.ar(\effectReverbOut.kr(0), snd*\effectReverb.kr(0));
    	}).add;

    	SynthDef("MiPlaits",{
    		var ducked;

    		var env = EnvGen.ar(
    			Env.adsr(
    				\attack.kr(0.1),
    				\decay.kr(1.0),
    				\sustain.kr(0.7),
    				\release.kr(0.5)
    			),
    			1-EnvGen.kr(Env.new([0,0,1],[\duration.kr(1),0])),
    			doneAction:2,
    		);
    		var snd = MiPlaits.ar(
    			pitch: \note.kr(60).min(127).max(0),
    			engine: \engine.kr(0).min(15).max(0),
    			harm: \engine.kr(0),
    			timbre: \timbre.kr(0),
    			morph: \morph.kr(0),
    			trigger: 1,
    			fm_mod: \fm_mod.kr(0),
    			timb_mod: \timb_mod.kr(0),
    			morph_mod: \morph_mod.kr(0),
    			lpg_colour: \lpg_colour.kr(0),
    		);
    		snd = snd * env * \trackVolume.kr(0).dbamp * (1.0 / \noteSize.kr(1).sqrt)
    						* \velocity.kr(100).min(127).max(0).linlin(0,127,-24,24).dbamp;
    		snd = RLPF.ar(snd, \lowPassFilter.kr(20000), 0.707);
    		snd = Balance2.ar(snd[0],snd[1],\pan.kr(0));

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);

    		Out.ar(\trackOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectDryOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectCombOut.kr(0), snd*\effectComb.kr(0));
    		Out.ar(\effectReverbOut.kr(0), snd*\effectReverb.kr(0));
    	}).add;

    	SynthDef("MiBraids",{
    		var ducked;
    		var env = EnvGen.ar(
    			Env.adsr(
    				\attack.kr(0.1),
    				\decay.kr(1.0),
    				\sustain.kr(0.7),
    				\release.kr(0.5)
    			),
    			1-EnvGen.kr(Env.new([0,0,1],[\duration.kr(1),0])),
    			doneAction:2,
    		);
    		var snd = MiBraids.ar(
    			pitch: \note.kr(60),
    			timbre: \timbre.kr(0),
    			color: \color.kr(0),
    			model: \model.kr(0),
    			trig: 1,
    			resamp: \resamp.kr(0),
    			decim: \decim.kr(1),
    			bits: \bits.kr(0),
    			ws: \ws.kr(0),
    		);
    		snd = snd * env * \trackVolume.kr(0).dbamp * (1.0 / \noteSize.kr(1).sqrt)
    			* \velocity.kr(100).min(127).max(0).linlin(0,127,-24,24).dbamp;
    		snd = RLPF.ar(snd, \lowPassFilter.kr(20000), 0.707);

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);

    		snd = Pan2.ar(snd,\pan.kr(0));
    		Out.ar(\trackOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectDryOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectCombOut.kr(0), snd*\effectComb.kr(0));
    		Out.ar(\effectReverbOut.kr(0), snd*\effectReverb.kr(0));
    	}).add;

    	SynthDef("PolyPerc",{
    		var ducked;
    		var env = EnvGen.ar(
    			Env.adsr(
    				\attack.kr(0.1),
    				\decay.kr(1.0),
    				\sustain.kr(0.7),
    				\release.kr(0.5)
    			),
    			1-EnvGen.kr(Env.new([0,0,1],[\duration.kr(1),0])),
    			doneAction:2,
    		);
    		var snd = Pulse.ar(\note.kr(60).midicps, 0.5);
    		snd = MoogFF.ar(snd, \lowPassFilter.kr(20000), \resonance.kr(1.5));
    		snd = snd * env * \trackVolume.kr(0).dbamp * (1.0 / \noteSize.kr(1).sqrt)
    			* \velocity.kr(100).min(127).max(0).linlin(0,127,-24,0).dbamp;

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);

    		snd = Pan2.ar(snd,\pan.kr(0));
    		Out.ar(\trackOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectDryOut.kr(0), snd*(1.0 - \effectReverb.kr(0)));
    		Out.ar(\effectCombOut.kr(0), snd*\effectComb.kr(0));
    		Out.ar(\effectReverbOut.kr(0), snd*\effectReverb.kr(0));
    	}).add;

    	SynthDef("externalInput", {
    		arg trackId,
    		trackVolume = 0,
    		pan=0,
    		inbus=0,
    		trackOut,
    		effectDryOut,
    		effectReverb = 0.0, effectReverbOut,
    		effectComb = 0.0, effectCombOut
    		;
    		var snd, ducked;
    		snd = SoundIn.ar([0,1]) * EnvGen.ar(Env.adsr(1.0,0.0,1.0,1.0),1) * DC.ar(0); // TODO:remove 0
    		snd = Balance2.ar(snd[0], snd[1], pan);
    		snd = snd * -10.dbamp * trackVolume.dbamp;

    		// check if signal is ducked
    		// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    		ducked = Compander.ar(
    			in:         snd,     // signal to be ducked
    			control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    			thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    			slopeBelow: 1,       // unchanged below threshold
    			slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    			clampTime:  \duckingAttack.kr(0.02),     // attack
    			relaxTime:  \duckingRelease.kr(0.20)      // release
    		);
    		// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    		snd = Select.ar(\duckingType.kr(0),
    			[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    		);
    		// tiny floor to avoid denormals at extreme reductions:
    		snd = snd + (1e-6 * WhiteNoise.ar(1));
    		Out.ar(\duckingBusOut.kr(0),
    			Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    				0,  // 0 = not ducking
    				1,  // 1 = is ducking (writer)
    				0   // 2 = not ducking
    			])
    		);


    		Out.ar(trackOut, snd*(1.0 - effectReverb));
    		Out.ar(effectDryOut, snd*(1.0 - effectReverb));
    		Out.ar(effectCombOut, snd*effectComb);
    		Out.ar(effectReverbOut, snd*effectReverb);
    	}).add;

    	SynthDef("diskout", { arg bufnum=0, inbus=0, gate=1;
    		var snd = In.ar(inbus,2);
    		snd = snd * EnvGen.ar(Env.adsr(0.001,0.0,1.0,1.0),gate,doneAction:2);
    		DiskOut.ar(bufnum,snd);
    	}).add;

    	2.do({
    		arg ch;
    		SynthDef("playback"++(ch+1),{
    			arg out,buf,gate = 1;
    			var env = EnvGen.ar(Env.adsr(0.01,0.0,1.0,0.2),gate,doneAction:2);
    			var snd = PlayBuf.ar(ch+1, buf, BufRateScale.kr(buf), loop: 0, doneAction:2);
    			if (ch<1,{
    				snd = Pan2.ar(snd,0);
    			});
    			Out.ar(out, (snd*env));
    		}).add;
    	});

    	2.do({
    		arg ch;
    		SynthDef("sampler"++(ch+1),{
    			arg buf,
    			trackId=0,
    			volumeDB=0,
    			rate = 1.0,
    			pitch = 0.0,
    			gate = 1,
    			t_trig = 1,
    			xfade=0.01,
    			bpmSource=120,
    			bpmTarget=120,
    			lowPassFilter=20000,
    			highPassFilter=20,
    			trackVolume = 0.0,
    			pan=0,
    			// retriggering
    			finalVolumeToStart=0,
    			retrigNumTotal=0,
    			retrigRateChangeBeats=1,
    			retrigRateStart=1,
    			retrigRateEnd=0,
    			retrigPitchChange=0,
    			finalPitchToStart=0,
    			retrigVolumeChange=0,
    			sliceAttackBeats = 0.001,
    			sliceDurationBeats = 0,
    			sliceReleaseBeats = 0.001,
    			sliceCount = 32, // number of slices to cut the sample into
    			sliceStart = 0.0, // slice start position (0.0-1.0, calculated in Go)
    			sliceEnd = 1.0,   // slice end position (0.0-1.0, calculated in Go)
				sliceBounce = 0.0,
				sliceStop = 0.0,
    			trackOut,
    			effectDry = 1.0,
    			effectDryOut,
    			effectComb = 0.0,
    			effectCombOut,
    			effectDelay = 0.0,
    			effectReverse = 0,
    			effectTimestretchStart = 0,
    			effectTimestretchEnd = 0,
    			effectTimestretchBeats = 0,
    			effectReverb = 0.0,
    			effectLPFStart = 20000,
    			effectLPFEnd = 0,
    			effectLPFBeats = 0,
    			effectReverbOut
    			;
    			var snd,sndA,sndB,crossfade,aOrB;
    			var sndDry, sndReverb, sndComb;
    			var posA,posB,posCur,posCurFeedback,posCurFeedbackLast;
    			var retrigTrig, retrigCount;
    			var frames=BufFrames.ir(buf);
    			var syncBpm=(\synctobpm.ir(0) * bpmTarget/bpmSource) + (1 - \synctobpm.ir(0));
    			var seconds=BufDur.ir(buf) / syncBpm;
    			var secondsLeft = seconds;
    			var sliceSeconds = seconds / sliceCount;
    			var pos = 0.0;
    			var sliceTrigger = t_trig;
    			var beatDuration = 60 / bpmTarget;
    			var retrigCountFeedback = 0;
    			var timestretchPos, timestretchRate, effectTimestretch;
    			var side, atk, rel, depth, slopeAbove, thresh, ducked;
				var sliceStartOriginal = sliceStart;
				var sliceEndOriginal = sliceEnd;
				var localInData;
				var endToggled;
    			// Position is always provided by Go code (sliceStart and sliceEnd)
    			// Handle reverse playback by swapping start and end
				
    			pos = Select.kr(effectReverse>0,[
    				sliceStart,  // forward: use start position
    				sliceEnd     // reverse: use end position
    			]);

    			// compute seconds left
    			secondsLeft = Select.kr(effectReverse > 0, [
    				(seconds * (1 - pos)), // forward: seconds to end
    				(seconds * pos)      // reverse: seconds to beginning
    			]);

    			// set effectLPFEnd to effectLPFStart if it is 0
    			effectLPFEnd = Select.kr(effectLPFEnd < 0.001, [
    				effectLPFEnd,
    				effectLPFStart,
    			]);
    			// set effectTimestretch to beginning if it is 0
    			effectTimestretchEnd = Select.kr(effectTimestretchEnd< 0.001,[
    				effectTimestretchEnd,
    				effectTimestretchStart,
    			]);
    			effectTimestretch = Line.kr(effectTimestretchStart, effectTimestretchEnd, effectTimestretchBeats*beatDuration);
    			timestretchRate = Select.ar(effectTimestretch>1,[
    				DC.ar(pos*frames),
    				Phasor.ar(trig:TDelay.ar(Impulse.ar(0, 0),0),rate:rate/effectTimestretch,start:0,end:frames,resetPos:(pos*frames)).floor,
    			]);


    			// make sure retrigRateEnd is at least retrigRateStart if its 0
    			retrigRateEnd = Select.kr(retrigRateEnd < 0.001, [retrigRateEnd, retrigRateStart]);

    			// if sliceDurationBeats = 0, make it infinite
    			sliceDurationBeats = Select.kr(sliceDurationBeats < 0.001, [sliceDurationBeats, (seconds/sliceCount)/(60/bpmSource)]);

    			// Calculate rate
    			rate = rate*BufRateScale.ir(buf)*syncBpm;
    			// Calculate pitch
    			pitch = pitch + (retrigNumTotal * retrigPitchChange.neg * finalPitchToStart);
    			rate = rate * (2 ** (pitch/12.0));

    			// Calculate retriggers
    			localInData = LocalIn.ar(3);
				retrigCountFeedback = localInData[0];
				posCurFeedback = localInData[1];
				posCurFeedbackLast = localInData[2];
    			retrigTrig = Impulse.ar(XLine.ar(retrigRateStart,retrigRateEnd,beatDuration*retrigRateChangeBeats) / beatDuration);
    			retrigCount = PulseCount.ar(retrigTrig * (retrigCountFeedback<retrigNumTotal));
    			sliceTrigger = sliceTrigger + (retrigTrig*(retrigCount<retrigNumTotal));

    			// Increase pitch based on retrigger
    			// rate = rate * (2 ** (retrigCount / 12));
    			// Lower pitch based on retrigger
    			rate = rate * (2 ** (retrigPitchChange * (retrigCount / 12)));
    			rate = rate.abs;
    			// Do reverse
    			rate = Select.kr(effectReverse>0,[
    				rate,
    				rate * -1,
    			]);

				endToggled = ToggleFF.kr(
					((posCurFeedback > (sliceEndOriginal - 0.001)) * (posCurFeedbackLast <= (sliceEndOriginal - 0.001)))
					+
					((posCurFeedback < (sliceStartOriginal + 0.001)) * (posCurFeedbackLast >= (sliceStartOriginal + 0.001)))
				);
				// bounce if end toggled
				rate = rate * Select.kr(endToggled * sliceBounce,[1.0,-1.0]);
				gate = gate * Select.kr(endToggled * (1 - sliceBounce),[1,0]);
				


    			// Change volume based on retrigger
    			volumeDB = volumeDB + (A2K.kr(retrigCount) * retrigVolumeChange) + (retrigVolumeChange.neg * retrigNumTotal * finalVolumeToStart);


    			// do timestretching
    			sliceTrigger = sliceTrigger + ((effectTimestretch>1)*Impulse.ar(effectTimestretch/sliceSeconds));


    			// Determine whether to toggle playback
    			aOrB=ToggleFF.ar(sliceTrigger);
    			crossfade=VarLag.ar(K2A.ar(aOrB),xfade,warp:\sine);

    			posA=Phasor.ar(
    				trig:(1-aOrB),
    				rate:rate,
    				end:frames,
    				resetPos:timestretchRate,
    			);
    			posB=Phasor.ar(
    				trig:aOrB,
    				rate:rate,
    				end:frames,
    				resetPos:timestretchRate,
    			);
				posCur = Select.ar(aOrB,[posA,posB])/frames;
    			snd=(BufRd.ar(
    				numChannels:ch+1,
    				bufnum:buf,
    				phase:posA,
    			)*crossfade)+(BufRd.ar(
    				numChannels:ch+1,
    				bufnum:buf,
    				phase:posB,
    			)*(1-crossfade));

    			snd = snd * Lag.kr(volumeDB.dbamp,0.2);

    			// envelope
    			snd = snd *
    				(
    					// playthrough = 0, sliced mode
    					((1 - \playthrough.ir(0)) *
    					EnvGen.ar(Env.new([0,1,1,0],[sliceAttackBeats,sliceDurationBeats,sliceReleaseBeats]*beatDuration,[-4,4]), sliceTrigger))
    					// playthrough = 1, playthrough mode
    					+ (\playthrough.ir(0) *
    					EnvGen.ar(Env.new([0,1,1,0],[0.001,secondsLeft-(0.005),0.001],[-4,4]), 1, doneAction:2 * \playthrough.ir(0))
    					)
    				);


    			// full envelope
    			snd = snd * EnvGen.ar(Env.adsr(xfade,0.0,1.0,xfade,curve:\sine),gate,doneAction:2);

    			// filtering
    			snd = RLPF.ar(snd, XLine.kr(effectLPFStart, effectLPFEnd, effectLPFBeats*beatDuration), 0.303);
    			DetectSilence.ar(snd,time:1,doneAction:2);

    			// filtering
    			snd = RLPF.ar(snd, lowPassFilter, 0.707);
    			snd = RHPF.ar(snd, highPassFilter, 0.707);

    			// panning
    			if (ch<1,{
    				snd = Pan2.ar(snd,pan);
    			},{
    				snd = Balance2.ar(snd[0],snd[1],pan);
    			});

    			// volume
    			snd = snd * trackVolume.dbamp * \velocity.kr(100).min(127).max(0).linlin(0,127,-24,24).dbamp;

    			// check if signal is ducked
    			// process: when the sidechain exceeds thresh, reduce 'snd' by slopeAbove
    			ducked = Compander.ar(
    				in:         snd,     // signal to be ducked
    				control:    LeakDC.ar(In.ar(\duckingBusIn.kr(0), 1)),    // external sidechain
    				thresh:     \duckingThresh.kr(0.02),  // start reducing above this
    				slopeBelow: 1,       // unchanged below threshold
    				slopeAbove: (1 - \duckingDepth.kr(0).clip(0, 0.99)).max(0.01),
    				clampTime:  \duckingAttack.kr(0.02),     // attack
    				relaxTime:  \duckingRelease.kr(0.20)      // release
    			);
    			// engage only for ducked voices (duckingType == 2). Writers and non-duckers pass through unchanged.
    			snd = Select.ar(\duckingType.kr(0),
    				[snd, snd, ducked]   // 0:not ducked, 1:writer, 2:ducked(reader)
    			);
    			// tiny floor to avoid denormals at extreme reductions:
    			snd = snd + (1e-6 * WhiteNoise.ar(1));
    			Out.ar(\duckingBusOut.kr(0),
    				Mix(snd) * \duckingDepth.kr(0).clip(0, 0.99) * Select.kr(\duckingType.kr(0), [
    					0,  // 0 = not ducking
    					1,  // 1 = is ducking (writer)
    					0   // 2 = not ducking
    				])
    			);

				// Send out the slice start/stop and playhead position and gate info for the track
				SendReply.kr(Impulse.kr(30),'/sampler_playhead',[
					trackId,
					gate,
					posCur, 
					sliceStartOriginal,
					sliceEndOriginal
				]);

    			// Feedback
    			LocalOut.ar([retrigCount,posCur, posCurFeedback]);
    			Out.ar(trackOut, snd*(1.0 - effectReverb));
    			Out.ar(effectDryOut, snd*effectDry);
    			Out.ar(effectCombOut, snd*XLine.kr(0.001,effectComb,sliceDurationBeats*beatDuration/2));
    			Out.ar(effectReverbOut, snd*XLine.kr(0.001,effectReverb,sliceDurationBeats*beatDuration/2));
    		}).add;
    	});

    	SynthDef("startupSound",{
    		var snd;
    		snd = SinOsc.ar([440,442],0,0.5) * EnvGen.ar(Env.perc(0.01,4.0),1,doneAction:2);
    		snd = snd + (
    			SinOsc.ar([660,665],0,0.2) * EnvGen.ar(Env.perc(0.01,4.0),TDelay.kr(DC.kr(1.0),1.0),doneAction:2)
    		);
    		Out.ar(\out.kr(0), snd);
    	}).add;

    	SynthDef("out",{
    		arg busReverb, busDry, busComb, busDisk,
    		volumeDB=0.0,
    		reverbAmt=0.1,
    		pregain=0.0,
    		postgain=0.0,
    		tape=0.0,
    		bias=6.0.neg,
    		saturation=6.0.neg,
    		drive=6.0.neg,
    		shimmer=1.0,
    		combAmt=0.0,
    		track0Bus,
    		track1Bus,
    		track2Bus,
    		track3Bus,
    		track4Bus,
    		track5Bus,
    		track6Bus,
    		track7Bus,
    		track8Bus; // track "8" is the external input
    		var sndWet = In.ar(busReverb,2);
    		var sndDry = In.ar(busDry,2);
    		var sndComb = In.ar(busComb,2);
    		var snd = 				sndDry;
    		SendReply.kr(Impulse.kr(30),'/track_volume',[Lag.kr(Amplitude.kr([
    			Mix.new(In.ar(track0Bus,2)),
    			Mix.new(In.ar(track1Bus,2)),
    			Mix.new(In.ar(track2Bus,2)),
    			Mix.new(In.ar(track3Bus,2)),
    			Mix.new(In.ar(track4Bus,2)),
    			Mix.new(In.ar(track5Bus,2)),
    			Mix.new(In.ar(track6Bus,2)),
    			Mix.new(In.ar(track7Bus,2)),
    			Mix.new(In.ar(track8Bus,2)),
    		],0.3,0.3).max(0.00001).ampdb,3)]);
			// Send out /track_waveform message with the normalized waveform of each track 
			SendReply.kr(Impulse.kr(30),'/track_waveform',[
				Normalizer.ar(LPF.ar(In.ar(track0Bus,2)[0],60))*(Amplitude.kr(In.ar(track0Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track1Bus,2)[0],60))*(Amplitude.kr(In.ar(track1Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track2Bus,2)[0],60))*(Amplitude.kr(In.ar(track2Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track3Bus,2)[0],60))*(Amplitude.kr(In.ar(track3Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track4Bus,2)[0],60))*(Amplitude.kr(In.ar(track4Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track5Bus,2)[0],60))*(Amplitude.kr(In.ar(track5Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track6Bus,2)[0],60))*(Amplitude.kr(In.ar(track6Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track7Bus,2)[0],60))*(Amplitude.kr(In.ar(track7Bus,2)[0]).ampdb>70.neg),
				Normalizer.ar(LPF.ar(In.ar(track8Bus,2)[0],60))*(Amplitude.kr(In.ar(track8Bus,2)[0]).ampdb>70.neg),
			]);

    		// add in comb
    		snd = snd + ((0.5*sndComb)+
    			Pan2.ar(CombN.ar(sndComb[0], 0.02, Rand(0.01,0.06), Rand(1,2)),-1,0.9) +
    			Pan2.ar(CombN.ar(sndComb[1], 0.02, Rand(0.01,0.06), Rand(1,2)),1,0.9));

    		// add in reverb
    		sndWet = DelayN.ar(sndWet, 0.03, 0.03);
    		sndWet = sndWet + PitchShift.ar(sndWet, 0.13, 2,0,1,1*shimmer/2);
    		sndWet = sndWet + PitchShift.ar(sndWet, 0.1, 4,0,1,0.5*shimmer/2);
    		sndWet = sndWet + PitchShift.ar(sndWet, 0.1, 8,0,1,0.125*shimmer/2);
    		snd = snd + Fverb.ar(sndWet[0],sndWet[1],200,
    			tail_density: LFNoise2.ar(1/3).range(50,90),
    			decay: LFNoise2.ar(1/3).range(50,90),
    		);

    		snd = RHPF.ar(snd,60,0.303);
    		snd = snd * Lag.kr(pregain).dbamp;
    		snd = SelectX.ar(Lag.kr(tape),[snd,AnalogTape.ar(snd,
    			bias: Lag.kr(bias).dbamp,
    			saturation: Lag.kr(saturation).dbamp,
    			drive: Lag.kr(drive).dbamp,
    			oversample: 2,
    		)]);
    		snd = snd * Lag.kr(volumeDB).dbamp * Lag.kr(postgain).dbamp;

    		SendReply.kr(Impulse.kr(30),'/waveform',Normalizer.ar(LPF.ar(snd[0],60))*(Amplitude.kr(snd[0]).ampdb>70.neg));
    		ReplaceOut.ar(0,snd);
    		Out.ar(busDisk, snd);
    	}).add;

    	s.sync;
    	~busDry = Bus.audio(s, 2);
    	~busReverb = Bus.audio(s, 2);
    	~busComb = Bus.audio(s, 2);
    	~busDisk = Bus.audio(s, 2);
    	~busTrack = Array.fill(9, { Bus.audio(s, 2) });
    	~busDucking = Array.fill(9, { Bus.audio(s, 1) });
    	~grpDuckWrite = Group.head(Server.default);
    	~grpDuckRead  = Group.after(~grpDuckWrite);
    	~grpFX = Group.after(~grpDuckRead);
    	s.sync;
    	~synOut = Synth.tail(~grpFX,"out",[
    		busReverb: ~busReverb,
    		busDry: ~busDry,
    		busComb: ~busComb,
    		busDisk: ~busDisk,
    		track0Bus: ~busTrack[0],
    		track1Bus: ~busTrack[1],
    		track2Bus: ~busTrack[2],
    		track3Bus: ~busTrack[3],
    		track4Bus: ~busTrack[4],
    		track5Bus: ~busTrack[5],
    		track6Bus: ~busTrack[6],
    		track7Bus: ~busTrack[7],
    		track8Bus: ~busTrack[8],
    		volumeDB: -24,
    	]);
    	s.sync;
    	~synthsPlaying.put(8, Dictionary.new());
    	~synthsPlaying.at(8).put(0, Synth.head(Server.default,"externalInput",[
    		inbus: 0,
    		trackOut: ~busTrack[8],
    		effectDryOut: ~busDry,
    		effectReverbOut: ~busReverb,
    		effectCombOut: ~busComb,
    		trackId: 8,
    		trackVolume: 0,
    		pan: 0,
    	]));
    	NodeWatcher.register(~synthsPlaying.at(8).at(0));
    	s.sync;
    	~dx7syn = thisProcess.interpreter.executeFile(
    		PathName(thisProcess.nowExecutingPath).pathOnly +/+ "DX7.scd"
    	);
    	s.sync;
    	~sampleCache = Dictionary.new();


    	~playSynthFromMsg = {
    		arg msg;
    		var synName = 1000000.rand.asString;
    		var track = msg[1].asInteger;
    		var argLast;
    		var noteOn = msg[2].asInteger;
    		var synthToPlay = msg[3].asString;
    		var notes = Array.new();
    		var nonNoteIndex = 4;
    		var keepSearching = true;
			var polyphonic = true;
			var monophonic = false;
			var lastMessage = nil;
			msg.do({ arg item, i;
				if (lastMessage==\monophonic,{
					polyphonic = (item<1);
					monophonic = (item>0);
				});
				lastMessage = item;
			});
    		// find where msg[4:] is not a float
    		msg[4..].do({ |v,i|
    			if (v.isNumber && keepSearching,{
    				notes = notes.add(v);
    				nonNoteIndex = i + 5;
    			},{
    				// break out of do loop
    				keepSearching = false;
    			});
    		});
    		["track",track,"noteOn",noteOn,"notes",notes].postln;
    		// create a new dictionary if it doesn't exist
    		if (~synthsPlaying.at(track).isNil,{
    			~synthsPlaying.put(track, Dictionary.new());
    		});

    		// stop all currently playing synths for this track
			if (polyphonic, {
				~synthsPlaying.at(track).values.do({
					arg syn;
					if (syn.isPlaying, {
						if (syn.notNil,{
							[syn,"stopped"].postln;
							syn.set(\gate,0);
						});
					});
				});
			});
    		if (noteOn > 0,{
    			var args;
    			var dict = Dictionary.new;
    			var targetGroup = ~grpDuckRead;
    			dict.putAll((
    				buf:             b,
    				effectDryOut:    ~busDry,
    				effectCombOut:   ~busComb,
    				effectReverbOut: ~busReverb,
    				trackId:         track,
    				trackOut:        ~busTrack[track],
					t_trig:			1,
    			));
    			dict.putPairs(msg.copyToEnd(nonNoteIndex));
    			// round numbers to 1/128 in place
    			dict.keysValuesChange { |k, v|
    				[k,v].postln;
    				if (v.isNumber) { v.round(1.0/128.0) } { v }
    			};

    			// rewrite ducking bus
    			if (dict.includesKey(\duckingBus),{
    				var i = dict[\duckingBus].asInteger;
    				dict.put(\duckingBusIn,  ~busDucking[i]);
    				dict.put(\duckingBusOut, ~busDucking[i]);
    				dict.removeAt(\duckingBus);
    			},{
    				// nothing happens on bus 8
    				dict.put(\duckingBusIn, ~busDucking[8]);
    				dict.put(\duckingBusOut, ~busDucking[8]);
    			});
    			if (dict.includesKey(\duckingType), {
    				if (dict[\duckingType] == 1, {
    					targetGroup = ~grpDuckWrite;
    				});
    			});
    			args = dict.asPairs;

    			notes.do({ arg n;
    				var synthArgs = args ++ [\note,n,\noteSize,notes.size];
    				var synthName = synName ++ "_" ++ n.asString;
    				// print out all the synthargs
    				synthArgs.do({ arg a,i; [i,a].postln; });
					// if monophonic check to see if there is a synth playing just use that instead
					if (monophonic, {
						var playingSynth = ~synthsPlaying.at(track).values.detect { |syn| syn.notNil and: { syn.isPlaying } };
						["monophonic", playingSynth].postln;	
						if (playingSynth.notNil, {
							["reused", playingSynth].postln;
							playingSynth.set(*synthArgs);
							^nil;
						});
					});
					// play new synth
					["playing",synthName].postln;
    				~synthsPlaying.at(track).put(synthName,
    					Synth.head(targetGroup,synthToPlay,synthArgs).onFree({
    						[synthName,"freed"].postln;
    						~synthsPlaying.at(track).removeAt(synthName);
    					});
    				);
    				NodeWatcher.register(~synthsPlaying.at(track).at(synthName));
    			});
    		});
    	};


    	~playFromMsg = {
    		arg msg, b;
    		var synName = 1000000.rand.asString;
    		var track = msg[2].asInteger;
    		var argLast;
    		var dict = Dictionary.new;
    		var targetGroup = ~grpDuckRead;
    		dict.putAll((
    		    buf:             b,
    		    effectDryOut:    ~busDry,
    		    effectCombOut:   ~busComb,
    		    effectReverbOut: ~busReverb,
    		    trackId:         track,
    		    trackOut:        ~busTrack[track]
    		));
    		dict.putPairs(msg.copyToEnd(3));
    		// round numbers to 1/128 in place
    		dict.keysValuesChange { |k, v|
    			[k,v].postln;
    		    if (v.isNumber) { v.round(1.0/128.0) } { v }
    		};

    		// rewrite ducking bus
    		if (dict.includesKey(\duckingBus),{
    			var i = dict[\duckingBus].asInteger;
    			dict.put(\duckingBusIn,  ~busDucking[i]);
    			dict.put(\duckingBusOut, ~busDucking[i]);
    			dict.removeAt(\duckingBus);
    		},{
    			// nothing happens on bus 8
    			dict.put(\duckingBusIn, ~busDucking[8]);
    			dict.put(\duckingBusOut, ~busDucking[8]);
    		});
    		if (dict.includesKey(\duckingType), {
    			if (dict[\duckingType] == 1, {
    				targetGroup = ~grpDuckWrite;
    			});
    		});


    		// create a new dictionary for the track if it doesn't exist
    		if (~samplesPlaying.at(track).isNil,{
    			~samplesPlaying.put(track, Dictionary.new());
    		});

    		if (
    		    (dict.includesKey(\update).not) or: { dict[\update] == 0 }
    		) {
    		    // stop all synths
    		    ~samplesPlaying.at(track).values.do { |syn|
    		        if (syn.notNil and: { syn.isPlaying }) {
    		            syn.set(\gate, 0);
    		        }
    		    };
    		    // play new synth
    		    ~samplesPlaying.at(track).put(synName,
    		        Synth.head(targetGroup, "sampler" ++ (b.numChannels), dict.asPairs).onFree({
    		            [b, "freed"].postln;
    		            ~samplesPlaying.at(track).removeAt(synName);
    		        })
    		    );
    		    ["played",~samplesPlaying.at(track).at(synName)].postln;
    		    NodeWatcher.register(~samplesPlaying.at(track).at(synName));
    		} {
    		    // set all synths
    		    ~samplesPlaying.at(track).values.do { |syn|
    		        if (syn.notNil and: { syn.isPlaying }) {
    		        	["updating",syn].postln;
    		            syn.set(*dict.asPairs);
    		        }
    		    };
    		};
    	};

    	OSCFunc({ |msg|
    		var filename = msg[1];
    		msg.postln;
			// log every key/value pair
			msg.do({ |v,i|
				[i,v].postln;
			});
    		if (~sampleCache.at(filename).isNil,{
    			// load sample
    			~sampleCache.put(filename, Buffer.read(s,filename,action:{ |b|
    				["loaded",b,b.numChannels].postln;
    				~playFromMsg.(msg,b);
    			}));
    		},{
    			~playFromMsg.(msg,~sampleCache.at(filename));
    		});
    	},'/sampler');
    	OSCFunc({ |msg|
    		var synthToPlay = msg[3].asString;
    		if (synthToPlay=="DX7",{
    			var settings = Dictionary.new();
    			var notes = Array.new();
    			var nonNoteIndex = 4;
    			var trackId = msg[1].asInteger;
    			var keepSearching = true;
    			// find where msg[4:] is not a float
    			msg[4..].do({ |v,i|
    				if (v.isNumber && keepSearching,{
    					notes = notes.add(v);
    					nonNoteIndex = i + 5;
    				},{
    					// break out of do loop
    					keepSearching = false;
    				});
    			});

    			settings.put("set",0);
    			settings.put("k","");
    			settings.put("v",0);
    			settings.put("preset",1959);
    			settings.put("synBefore",~synOut);
    			settings.put("note",60);
    			settings.put("vel",100);
    			settings.put("pan",0);
    			settings.put("attack",0.1);
    			settings.put("release",0.5);
    			settings.put("duration",1.0);
    			settings.put("trackVolume",-24.0);
    			settings.put("filter",20000);
    			settings.put("effectDryOut", ~busDry);
    			settings.put("effectCombOut", ~busComb);
    			settings.put("effectReverbOut", ~busReverb);
    			settings.put("trackOut", ~busTrack[trackId]);
    			settings.put("effectReverb",0.0);
    			settings.put("effectComb",0.0);
    			// for all other messages, they are key/value pairs
    			msg[nonNoteIndex..].do({ |v,i|
    				if (i.mod(2) == 1,{
    					settings.put(msg[nonNoteIndex + i - 1].asString,v);
    				});
    			});
    			// print all settings
    			settings.keysValuesDo({ |k,v|
    				[k,v].postln;
    			});
    			// normalize velocity (0-127) and add to track volume
    			if (settings.at("velocity").notNil,{
    				settings.put("trackVolume", settings.at("trackVolume") + settings.at("velocity").min(127).max(0).linlin(0,127,-24,24));
    			});
    			["playing DX7"].postln;
    			notes.do({ |n|
    				settings.put("note",n);
    				["note",n].postln;
    				~dx7syn.value(
    					settings.at("set"),
    					settings.at("k"),
    					settings.at("v"),
    					settings.at("preset"),
    					settings.at("synBefore"),
    					settings.at("note"),
    					settings.at("vel"),
    					settings.at("pan"),
    					settings.at("attack"),
    					settings.at("release"),
    					settings.at("duration"),
    					settings.at("trackVolume"),
    					settings.at("filter"),
    					settings.at("effectDryOut"),
    					settings.at("effectCombOut"),
    					settings.at("effectReverbOut"),
    					settings.at("effectReverb"),
    					settings.at("effectComb"),
    					settings.at("trackOut"),
    				);
    			});

    		},{
    			~playSynthFromMsg.(msg);
    		});
    	},'/instrument');

    	OSCFunc({ |msg|
    		// stop all currently playing synths in all tracks
    		~samplesPlaying.values.do({
    			arg track;
    			track.values.do({ arg syn;
    				if (syn.isPlaying,{
    					if (syn.notNil,{
    						syn.set(\gate,0);
    					});
    				});
    			});
    		});
    		// stop playback synth if it exists
    		if (~synthPlayback.notNil,{
    			if (~synthPlayback.isPlaying,{
    				[~synthPlayback,"stopped"].postln;
    				~synthPlayback.set(\gate,0);
    			});
    		});
    	},'/stop');
    	OSCFunc({ |msg|
    		NetAddr.new("127.0.0.1", 57121).sendMsg("/waveform", msg[3]);
    	},'/waveform');
		OSCFunc({ |msg|
			NetAddr.new("127.0.0.1", 57121).sendMsg("/sampler_playhead", *msg[3..].postln);
		},'/sampler_playhead');
    	OSCFunc({ |msg|
    		NetAddr.new("127.0.0.1", 57121).sendMsg("/track_volume", *msg[3..]);
    	},'/track_volume');
		OSCFunc({ |msg|
			NetAddr.new("127.0.0.1", 57121).sendMsg("/track_waveform", *msg[3..]);
    	},'/track_waveform');
    	OSCFunc({ |msg|
    		var filename = msg[1].asString;
    		var doRecord = msg[2].asInteger;
    		var enabledTracks = msg[3].asInteger;
    		var isTrackEnabled = { arg trackNum; (enabledTracks.bitAnd((2**trackNum).asInteger)) > 0 };
    		[filename,doRecord,enabledTracks,isTrackEnabled].postln;
    		// stop all recordings
    		["stopping all recordings"].postln;
    		~synthRecord.values.do({ |syn,name|
    			[syn,name,"stopping"].postln;
    			if (syn.notNil and: { syn.isPlaying }, {
    				[syn,"stopped"].postln;
    				syn.set(\gate,0);
    			});
    		});
    		~synthRecord.clear;
    		if (doRecord<1,{
    			// do nothing, recording already released
    		}, {
    			// start a recording
    			var recordingBuffer=Buffer.alloc(Server.default,65536,2);
    			var pathname=filename;
    			["starting recording",filename,"with track mask",enabledTracks.asBinaryString].postln;
    			recordingBuffer.write(pathname.standardizePath,PathName.new(pathname.standardizePath).extension,"int16",0,0,true);
    			~synthRecord.put(filename,Synth.tail(s,"diskout",[
    				\bufnum,recordingBuffer.bufnum,
    				\inbus,~busDisk,
    				\gate,1,
    			]).onFree({
    				[recordingBuffer,"freed"].postln;
    				recordingBuffer.free;
    			}));
    			NodeWatcher.register(~synthRecord.at(filename));
    			// create recorders only for enabled tracks (based on track mask)
    			9.do({ arg track;
    				var enabled = isTrackEnabled.(track);
    				if (enabled, {
    					var trackRecordingBuffer=Buffer.alloc(Server.default,65536,2);
    					var trackPathname=filename.asString.replace(".wav","_track"++track.asString++".wav");
    					["starting track recording for track",track,trackPathname].postln;
    					trackRecordingBuffer.write(trackPathname.standardizePath,PathName.new(trackPathname.standardizePath).extension,"int16",0,0,true);
    					~synthRecord.put("track"++track.asString,Synth.tail(s,"diskout",[
    						\bufnum,trackRecordingBuffer.bufnum,
    						\inbus,~busTrack[track],
    						\gate,1,
    					]).onFree({
    						[trackRecordingBuffer,"freed"].postln;
    						trackRecordingBuffer.free;
    					}));
    					NodeWatcher.register(~synthRecord.at("track"++track.asString));
    				}, {
    					["skipping track",track,"(not enabled in mask)"].postln;
    				});
    			});
    		});
    	},'/record');
    	OSCFunc({ |msg|
    		var filename = msg[1];
    		var gate = msg[2].asInteger;
    		msg.postln;
    		if (~synthPlayback.notNil,{
    			if (~synthPlayback.isPlaying,{
    				[~synthPlayback,"stopped"].postln;
    				~synthPlayback.set(\gate,0);
    			});
    		});
    		if (gate>0,{
    			Buffer.read(s,filename,action:{ |b|
    				~synthPlayback = Synth.head(Server.default,"playback"++(b.numChannels),[
    					\buf,b,
    					\out,~busDry,
    					\gate,1,
    				]).onFree({
    					[b,"freed"].postln;
    					b.free;
    				});
    				NodeWatcher.register(~synthPlayback);
    			});
    		});
    	},'/playback');
    	OSCFunc({ |msg|
    		["setting",msg[1],msg[2]].postln;
    		~synOut.set(msg[1],msg[2]);
    	},'/set');
    	OSCFunc({ |msg|
    		["/set_track",msg[1],msg[2],msg[3]].postln;
    		if (~synthsPlaying.at(msg[1].asInteger).notNil,{
    			~synthsPlaying.at(msg[1].asInteger).values.do({ arg syn;
    				if (syn.isPlaying,{
    					if (syn.notNil,{
    						[syn,"setting",msg[2],msg[3]].postln;
    						syn.set(msg[2].asString,msg[3]);
    					});
    				});
    			});
    		});
    	},'/set_track');

    	["loaded",NetAddr.langPort, NetAddr.localAddr].postln;

    	s.sync;
    	Routine {
    		inf.do({
    			NetAddr.new("127.0.0.1", 57121).sendMsg("/cpuusage", s.avgCPU);
    			1.sleep;
    		});
    	}.play;



    	//Server.default.record;
    	["playing startupSound"].postln;
    	// Synth.head(Server.default,"startupSound",[\out,~busDry]);
    	// [48,64,67,72].do({ |n|
    	// 	Synth.head(Server.default,"SuperSaw",[
    	// 		\note,n-12,\duration,4,\effectDryOut, ~busDry,\effectCombOut, ~busComb,\effectReverbOut, ~busReverb,
    	// 		\attack, 3.0,
    	// 		\release, 3.0,
    	// 		\sustain, 1.0,
    	// 		\effectReverb, 0.3,
    	// 		\lowPassFilter, 200,
    	// 		\lpa, 1,
    	// 		\lpenv, 7.0,
    	// 		\drive, 1.0,
    	// 		\noteSize, 4,
    	// 		\trackVolume, -12,
    	// 	]);
    	// });


		NetAddr.new("127.0.0.1", NetAddr.langPort).sendMsg(
			"/sampler",
			"/home/zns/Documents/collidertracker/beat_bpm174.wav",
			0,
			\trackVolume, 12.0,
			\sliceCount, 64,
			\bpmSource, 175.0,
			\bpmTarget, 171.0,
			\pitch, 0.0,
			\retrigNumTotal, 0,
			\retrigRateChangeBeats, 0.0,
			\retrigRateStart, 0.0,
			\retrigRateEnd, 0.0,
			\retrigPitchChange, 0.0,
			\retrigVolumeChange, 0.0,
			\finalPitchToStart, 0,
			\finalVolumeToStart, 0,
			\effectTimestretchStart, 0.0,
			\effectTimestretchEnd, 0.0,
			\effectTimestretchBeats, 0.0,
			\effectReverse, 0,
			\pan, 0.0,
			\lowPassFilter, 20000.0,
			\highPassFilter, 20.0,
			\effectComb, 0.0,
			\effectReverb, 0.0,
			\velocity, 64,
			\deltaTime, 0.17543859779835,
			\playthrough, 0,
			\synctobpm, 1,
			\sliceDurationBeats, 32,
			\sliceStart, 0,
			\sliceEnd, 0.125,
			\sliceBounce, 0,
			\sliceStop, 1,
		);

    }.play;

});
)
