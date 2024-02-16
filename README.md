__To build:__

```
% source env.bash
% make test
```

__To use:__

See the comments in [worksets/Makefile](worksets/Makefile).  We highlight a few recipes here.  We use the notation `LSV(d,q,f)` to
label the finite simplicial complex that is the quotient of an affine building as described in [EKZ] and [LSV].

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

__References__

[EKZ] Shai Evra, Tali Kaufman, Gilles Zémor. Decodable quantum LDPC codes beyond the $\sqrt{n}$ distance barrier using high dimensional expanders. \url{https://arxiv.org/abs/2004.07935v1}, 2020.

[LSV] Alexander Lubotzky, Beth Samuels, Uzi Vishne. Explicit constructions of Ramanujan complexes of type $\widetilde{A}_d$. Israel Journal of Mathematics, 149. 2005.
