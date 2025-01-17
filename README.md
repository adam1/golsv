__What is this?__

Short answer: mathematical tools

Technical answer: 

An LSV complex is a finite simplicial complex that is a quotient of an infinite simplicial complex known as an affine building of type $\tilde{A}_d$.  LSV complexes are one known method of constructing explicit examples of high-dimensional expanders (HDX), also known as Ramanujan expanders, and have been applied to quantum error correction [EKZ]. This repository contains tools created by the author to instantiate the smallest possible example of an LSV complex and measure its primary properties relevant to quantum error correction.  In particular, these tools generate what I call $L_2 = LSV(3,2,1+y+y^2)$ which is a two-dimensional complex with 60,480 vertices and 423,360 edges and triangles.  The tools compute the dimension of the first homology and the exact systole ($\dim H_1 = 19, \ S_1 = 8$). Exact calculation of the cosystole has eluded my efforts, but instead I implemented a coboundary decoder and sampled the error correction rate, which shows a logical error rate of 1% at a physical error rate of 10%, suggesting a cosystole of approximately $S^1 = \sim80,000$.  For a technical introduction, see the draft [here](https://reversible.io/TQC-2024-poster-summary.pdf) and the references below.

__Features__

The notation here refers to the draft linked above.

* Operations in the algebra $\mathcal{A}(R)$ and group $\mathcal{G}(R)$
* Projective matrix calculations in $\text{PGL}_3(\mathbb{F}_q)$ for $q = 4, 8, 16$
* Computing Cartwright-Steger generators for $\mathcal{G}(R)$ and their projective matrix representations
* Generation of Cayley graph for generators in $\mathcal{G}(R)$, $\mathcal{G}(R/I)$ and $\text{PGL}_3(\mathbb{F}_q)$
* Computing 2-d clique complex of a graph
* Representation of two-dimensional simplicial complexes (including graphs) as boundary matrices of $\mathbb{F}_2$-chain complexes
* Linear algebra with large sparse or small dense matrices over $\mathbb{F}_2$
* Matrix reduction aka computing Smith normal form 
* Computing dimension and generators for first homology (simplicial mod two homology) and cohomology
* Finding the systole of $L_2$ by computing the injectivity radius in the cover
* A simplicial cosystole algorithm (works for small complexes, too slow for $L_2$)


Most of the code involved directly in the above features is well-tested, but be warned there is plenty of experimental code here as well.  The reader is encouraged to email me if you want to use this code, and let's talk about it!

__To build__

```
source env.bash
make test
```

__To use__

See the comments in [worksets/Makefile](worksets/Makefile).  We highlight a few recipes here. 

**Recipe A.** Generate the Cayley complex of $L_2$ from scratch and compute the dimension of first homology.  This takes about one day on a contemporary laptop.

```
mkdir worksets/aaa
cd worksets/aaa
echo 111 > modulus.txt
echo 8 > max-depth.txt
make -f ../Makefile dim-H1.txt
```

**Recipe B.** Construct the boundary maps for the $L_4 = \text{LSV}(3,2,1+y+y^4)$.

```
mkdir worksets/bbb
cd worksets/bbb
make -f ../Makefile pgl-cayley-F16
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
