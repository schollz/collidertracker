Juno60 : MultiOutUGen {
    *ar { arg gate=0, note=60, amp=0.5,
          attack=0, decay=0.5, sustain=0.5, release=0.5,
          cutoff=0.5, resonance=0, envMod=0,
          saw=1, pulse=0, sub=0, noise=0, pwm=0.5,
          chorus=0, doneAction=0,
          lfoRate=0.5, lfoDelay=0.0, lfoAuto=0,
          dcoLfo=0.0, dcoPwmMod=0, dcoRange=1.0,
          hpf=0.0, vcfLfo=0.0, vcfKey=1.0, vcfDir=1.0,
          vcaType=0.0, vcaValue=0.5, chorusDrywet=1.0;

        ^this.multiNew('audio', gate, note, amp,
            attack, decay, sustain, release,
            cutoff, resonance, envMod,
            saw, pulse, sub, noise, pwm,
            chorus, doneAction,
            lfoRate, lfoDelay, lfoAuto,
            dcoLfo, dcoPwmMod, dcoRange,
            hpf, vcfLfo, vcfKey, vcfDir,
            vcaType, vcaValue, chorusDrywet)
    }
    
    *kr { arg gate=0, note=60, amp=0.5,
          attack=0, decay=0.5, sustain=0.5, release=0.5,
          cutoff=0.5, resonance=0, envMod=0,
          saw=1, pulse=0, sub=0, noise=0, pwm=0.5,
          chorus=0, doneAction=0,
          lfoRate=0.5, lfoDelay=0.0, lfoAuto=0,
          dcoLfo=0.0, dcoPwmMod=0, dcoRange=1.0,
          hpf=0.0, vcfLfo=0.0, vcfKey=1.0, vcfDir=1.0,
          vcaType=0.0, vcaValue=0.5, chorusDrywet=1.0;

        ^this.multiNew('control', gate, note, amp,
            attack, decay, sustain, release,
            cutoff, resonance, envMod,
            saw, pulse, sub, noise, pwm,
            chorus, doneAction,
            lfoRate, lfoDelay, lfoAuto,
            dcoLfo, dcoPwmMod, dcoRange,
            hpf, vcfLfo, vcfKey, vcfDir,
            vcaType, vcaValue, chorusDrywet)
    }
    
    checkInputs {
        ^this.checkValidInputs;
    }
    
    init { arg ... theInputs;
        inputs = theInputs;
        ^this.initOutputs(2, rate);
    }
}
