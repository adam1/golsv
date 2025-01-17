__What is this?__

Short answer: mathematical tools

Technical answer: 

An LSV complex is a finite simplicial complex that is a quotient of an infinite simplicial complex known as an affine building of type $\tilde{A}_d$ [LSV].  LSV complexes are one known method of constructing explicit examples of high-dimensional expanders (HDX), also known as Ramanujan expanders, and have been applied to quantum error correction [EKZ]. This repository contains tools created by the author to instantiate the smallest possible example of an LSV complex and measure its primary properties relevant to quantum error correction.  In particular, these tools generate what I call $L_2 = LSV(3,2,1+y+y^2)$ which is a two-dimensional complex with 60,480 vertices and 423,360 edges and triangles.  The tools compute the dimension of the first homology and the exact systole ($\dim H_1 = 19, \ S_1 = 8$). Exact calculation of the cosystole has eluded my efforts, but instead I implemented a coboundary decoder and sampled the error correction rate, which shows a logical error rate of 1% at a physical error rate of 10%, suggesting a cosystole of approximately $S^1 = \sim80,000$.

Most of the code involved directly in that effort is well-tested, but be warned there is plenty of experimental code here as well.  The reader is encouraged to email me if you want to use this code, and let's talk about it!

__To build__

```
source env.bash
make test
```

__To use__

See the comments in [worksets/Makefile](worksets/Makefile).  We highlight a few recipes here.  We use the notation `LSV(d,q,f)` to
identify the finite simplicial complex that is the quotient of an affine building as described in [EKZ] and [LSV].

**Recipe A.** Construct the `d1.txt` and `d2.txt` boundary maps for the LSV complex `LSV(3,2,1+y+y^4)`:

```
mkdir worksets/aaa
cd worksets/aaa
make -f ../Makefile/pgl-cayley-F16
```
And calculate the first homology:
```
make -f ../Makefile/dim-H1.txt
```

**Profiling**. Most of the programs are built with support for cpu profiling.  Profiling is off by default and can be toggled on and off by sending a `USR1` signal to the program.  Profile data is written to `cpu.out` and will be overwritten, so rename the files if you want to capture more than one profiling session in the lifetime of a process.

Example:
```
P=`ps -o comm,pid |grep smith |awk '{print $2}'`
kill -USR1 $P; sleep 600; kill -USR 1
go tool pprof -http : # opens browser
```

__References__

[EKZ] Shai Evra, Tali Kaufman, Gilles Zémor. Decodable quantum LDPC codes beyond the $\sqrt{n}$ distance barrier using high dimensional expanders. [https://arxiv.org/abs/2004.07935v1](https://arxiv.org/abs/2004.07935v1), 2020.

[LSV] Alexander Lubotzky, Beth Samuels, Uzi Vishne. Explicit constructions of Ramanujan complexes of type $\widetilde{A}_d$. Israel Journal of Mathematics, 149. 2005.
