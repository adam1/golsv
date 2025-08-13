all: dim-H1.txt

log=$(shell date '+%s').log
$(info log suffix is $(log))
$(shell touch $(log))

# We currently have two "Cayley expanders", programs that generate a
# Cayley graph/complex from generators.
#
# The first is the `cayley` program, which computes in the finite
# quotient (aka downstairs) using Cartwright-Steger generators as
# represented in PGL_3(F_4 or F_16).  These generators are taken from
# [LSV].  The resulting finite group in these cases is isomorphic to
# PGL(F_4) or PGL(F_16), respectively. Hence, this program computes
# Cay(PGL(F_4), S) or Cay(PGL(F_16), S) for the generating set S
# consisting of the Cartwright-Steger generators.
#
# The second (more recent) is the `calg-cayley` program, which
# computes in the original $\mathcal{G}(R)$ group over the
# Cartwright-Steger algebra (prior to the PGL matrix representation).
# It can compute either upstairs, in the infinite simplicial complex
# that is the affine building or downstairs, in the finite quotient.
#
# These two tools thus amount to different ways of computing a complex
# X, specifically producing the files:
#
#    d1.txt: matrix for boundary map d_1
#    d2.txt: matrix for boundary map d_2
#    vertex-basis.txt: basis for the vertex space (X_0)
#    edge-basis.txt: basis for the edge space (X_1)
#    triangle-basis.txt: basis for the triangle space (X_2)
#
# The subsequent linear algebra processing is common to both
# approaches.


# Tool 1.  This target computes simplicial complex
#
#     X= Cay(PGL(F_4), S)
#
# for the Cartwright-Steger generating set S, as encoded into the
# program.
pgl-cayley-F4:
	pgl-cayley -base-field F4 -genV2 -d1 d1.txt -d2 d2.txt -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt 2>&1 | tee cayley.$(log)

pgl-cayley-F8:
	pgl-cayley -base-field F8 -genV2 -d1 d1.txt -d2 d2.txt -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt 2>&1 | tee cayley.$(log)

pgl-cayley-F2:
	pgl-cayley -base-field F2 -genV2 -d1 d1.txt -d2 d2.txt -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt 2>&1 | tee cayley.$(log)

# This target corresponds to the section 10 example from [LSV].
pgl-cayley-F16:
	pgl-cayley -base-field F16 -genV2 -d1 d1.txt -d2 d2.txt -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt 2>&1 | tee cayley.$(log)

# Tool 2.  This tool also produces the basis and boundary files for a
# Cayley complex, but it does so by performing calculations in the
# Cartwright-Steger group.  It has separate modes for:
#
# 1. computing the Cayley graph (1-skeleton) and producing the vertex
#    and edge bases and d_1 boundary matrix
#
# 2. finding triangles (3-cliques) and producing the triangle basis
#    and d_2 boundary matrix
#
# 3. computing systolic candidates, producing the file S.txt,
#    whose columns are edge vectors.

calg-cayley-graph d1.txt vertex-basis.txt edge-basis.txt:
	calg-cayley -graph -quotient -modulus `cat modulus.txt` -max-depth `cat max-depth.txt` -d1 d1.txt \
		-vertex-basis vertex-basis.txt -edge-basis edge-basis.txt 2>&1 | tee calg-cayley-graph.$(log)

calg-cayley-fill-triangles d2.txt triangle-basis.txt: d1.txt
	calg-cayley -fill-triangles -quotient -modulus `cat modulus.txt` -max-depth `cat max-depth.txt` -d1 d1.txt -d2 d2.txt \
		-vertex-basis vertex-basis.txt -edge-basis edge-basis.txt \
		-triangle-basis triangle-basis.txt 2>&1 | tee calg-cayley-fill-triangles.$(log)

calg-cayley-systolic-candidates S.txt: vertex-basis.txt edge-basis.txt
	calg-cayley -modulus `cat modulus.txt` -max-depth `cat max-depth.txt` -d1 d1.txt -d2 d2.txt \
		-vertex-basis vertex-basis.txt -edge-basis edge-basis.txt \
		--systolic-candidates S.txt  2>&1 | tee calg-cayley-systolic-candidates.$(log)

gens-global.tex:
	calg-cayley -modulus `cat modulus.txt` -max-depth 1 -generators-latex-file gens-global.tex 2>&1 | tee gens-global.$(log)

gens-quotient.tex:
	calg-cayley -modulus `cat modulus.txt` -quotient -max-depth 1 -generators-latex-file gens-quotient.tex 2>&1 | tee gens-quotient.$(log)

# Many of the following targets can be used to compute homology from
# the d1.txt and d2.txt matrix files regardless of where they came
# from (i.e. they need not be produced by calg-cayley).

# This target computes a Smith normal form for the matrix d_1.
d1smith.txt d1colops.txt: d1.txt
	smith -in d1.txt -smith d1smith.txt -colops d1colops.txt 2>&1 | tee d1smith.$(log)

# This target computes a Smith normal form for the matrix d_2.
d2smith.txt d2colops.txt: d2.txt
	smith -in d2.txt -smith d2smith.txt -colops d2colops.txt 2>&1 | tee d2smith.$(log)

# This targets computes the dimension of Z_1, the kernel of d_1.
dim-Z1.txt: d1smith.txt
	dim -in d1smith.txt -kernel > dim-Z1.txt

# This targets computes the dimension of B_1, the image of d_2.
dim-B1.txt: d2smith.txt
	dim -in d2smith.txt -image > dim-B1.txt

# This target computes the dimension of H_1 = Z_1/B_1, the first homology group.
dim-H1.txt: dim-Z1.txt dim-B1.txt
	expr `cat dim-Z1.txt` - `cat dim-B1.txt` > dim-H1.txt

# This target computes a basis for Z_1.
Z1.txt: d1smith.txt d1colops.txt dim-Z1.txt
	automorphism -in d1colops.txt -dim `dim -in d1smith.txt -domain` -crop-start `dim -in d1smith.txt -image` -crop-end `dim -in d1smith.txt -domain` -out Z1.txt 2>&1 | tee Z1.$(log)

# This target checks that the basis for Z_1 we found is actually
# contained in the kernel of d_1.
check-Z1: d1Z1.txt
	rank=$$(dim -in d1Z1.txt -image) && if [[ $$rank -eq '0' ]]; then echo "Z1 is contained in ker d1"; else echo "Z1 is NOT contained in ker d1!"; false; fi

# This target multiplies matrices: d_1 * Z_1, producing d1Z1.txt,
# which should be all zeros.
d1Z1.txt: d1.txt Z1.txt
	multiply -A d1.txt -B Z1.txt -C d1Z1.txt 2>&1 | tee d1Z1.$(log)

# This target computes a basis for the coimage of d_2.
d2coimage.txt: d2smith.txt d2colops.txt
	automorphism -in d2colops.txt -dim `dim -in d2smith.txt -domain` -crop-start 0 -crop-end `dim -in d2smith.txt -image` -out d2coimage.txt 2>&1 | tee d2coimage.$(log)

# This target computes a basis for B_1, the image of d_2.
B1.txt: d2coimage.txt
	multiply -A d2.txt -B d2coimage.txt -C B1.txt 2>&1 | tee B1.$(log)

# This target computes a Smith normal form for the matrix B_1 (whose
# columns are a basis for B_1).
B1smith.txt B1rowops.txt B1colops.txt: B1.txt
	smith -in B1.txt -smith B1smith.txt -rowops B1rowops.txt -colops B1colops.txt 2>&1 | tee B1smith.$(log)

# The next three targets are used to compute an aligned basis for B_1
# inside of Z_1.  The resulting basis is stored as columns of matrix
# in U1.txt
P1T.txt: B1rowops.txt d1.txt
	automorphism -in B1rowops.txt -dim `dim -in d1.txt -domain` -out P1T.txt 2>&1 | tee P1T.$(log)

P1.txt: P1T.txt
	transpose -in P1T.txt -out P1.txt 2>&1 | tee P1.$(log)

U1.txt: B1smith.txt P1.txt B1colops.txt Z1.txt
	align -B1smith B1smith.txt -P P1.txt -B1colops B1colops.txt -Z1 Z1.txt -out U1.txt 2>&1 | tee U1.$(log)

# The following targets check that the basis for B_1 we found is
# actually contained in the image of d_2.
check-B1: KB1.txt
	rank=$$(dim -in KB1.txt -image) && if [[ $$rank -eq '0' ]]; then echo "B1 is contained in image d2"; else echo "B1 is NOT contained in image d2!"; false; fi

KB1.txt: K.txt B1.txt
	multiply -A K.txt -B B1.txt -C KB1.txt 2>&1 | tee KB1.$(log)

K.txt: KT.txt
	transpose -in KT.txt -out K.txt 2>&1 | tee K.$(log)

# Matrix KT has columns that form a basis for the kernel of d_2^T.
KT.txt: d2Tcolops.txt d2Tsmith.txt
	automorphism -in d2Tcolops.txt -dim `dim -in d2Tsmith.txt -domain` -crop-start `dim -in d2Tsmith.txt -image` -out KT.txt 2>&1 | tee KT.$(log)

d2Tsmith.txt d2Tcolops.txt: d2T.txt
	smith -in d2T.txt -smith d2Tsmith.txt -colops d2Tcolops.txt 2>&1 | tee d2Tsmith.$(log)

d2T.txt: d2.txt
	transpose -in d2.txt -out d2T.txt 2>&1 | tee d2T.$(log)

# Here we verify that columns of KT are in the kernel of d_2^T.
d2TKT.txt: d2T.txt KT.txt
	multiply -A d2T.txt -B KT.txt -C d2TKT.txt 2>&1 | tee d2TKT.$(log)

check-d2TKT: d2TKT.txt
	rank=$$(dim -in d2TKT.txt -image) && if [[ $$rank -eq '0' ]]; then echo "KT is contained in ker d2T"; else echo "KT is NOT contained in ker d2T!"; false; fi

# For completeness, verify that the columns of KT are independent.
check-KTsmith: KTsmith.txt
	rank=$$(dim -in KTsmith.txt -image) && if [[ $$rank -eq `dim -in KT.txt -domain` ]]; then echo "columns of KT are independent"; else echo "columns of KT are NOT independent!"; false; fi

KTsmith.txt: KT.txt
	smith -in KT.txt -smith KTsmith.txt 2>&1 | tee KTsmith.$(log)

# This target checks that the basis for U we found is actually
# contained in the kernel of d_1.
check-U: d1U1.txt KU1.txt
	rank=$$(dim -in d1U1.txt -image) && if [[ $$rank -eq '0' ]]; then echo "U1 is contained in ker d1"; else echo "U1 is NOT contained in ker d1!"; false; fi
	dim -in KU1.txt -check-cols-nonzero

d1U1.txt: d1.txt U1.txt
	multiply -A d1.txt -B U1.txt -C d1U1.txt 2>&1 | tee d1U1.$(log)

KU1.txt: K.txt U1.txt
	multiply -A K.txt -B U1.txt -C KU1.txt 2>&1 | tee KU1.$(log)

# This target searches for systolic vectors in Z_1 \setminus B_1.
systole.txt: U1.txt B1.txt
	systole -trials 1000 -B B1.txt -U U1.txt -min systole.txt 2>&1 | tee systole.$(log)

# Multiplying K by S gives a matrix in which each column corresponds
# to a systolic candidate and each coordinate within the column gives
# a dot product with the corresponding row of K, which is a generator
# for the kernel of d_2^T.  Hence if the column is all zeros, then the
# corresponding candidate is orthogonal to the kernel of d_2^T.  Thus
# by linear algebra, that candidate vector is in the image of d_2.
KS.txt: K.txt S.txt
	multiply -A K.txt -B S.txt -C KS.txt 2>&1 | tee KS.$(log)

Ssmith.txt: S.txt
	smith -in S.txt -smith Ssmith.txt 2>&1 | tee Ssmith.$(log)

rank-Ssmith.txt: Ssmith.txt
	dim -in Ssmith.txt -image > rank-Ssmith.txt

rank-KSsmith.txt: KSsmith.txt
	dim -in KSsmith.txt -image > rank-KSsmith.txt

KSsmith.txt: KS.txt
	smith -in KS.txt -smith KSsmith.txt 2>&1 | tee KSsmith.$(log)

# Verify that the systolic candidates are in the kernel of d_1.
d1S.txt: d1.txt S.txt
	multiply -A d1.txt -B S.txt -C d1S.txt 2>&1 | tee d1S.$(log)

check-d1S: d1S.txt
	rank=$$(dim -in d1S.txt -image) && if [[ $$rank -eq '0' ]]; then echo "candidates S are contained in ker d1"; else echo "candidates S are NOT contained in ker d1!"; false; fi

# WIP visualization of systolic cycles

# The matrix Sdiffs is a matrix in which each column corresponds to a 
# difference between a pair of systolic candidates.  The matrix is
# constructed by taking the difference of each pair of columns of S,
# ordered in the natural way.  Say S has n columns.  Then Sdiffs has
# n(n-1)/2 columns, ordered as follows: the first n-1 columns are the
# differences between the first column of S and each of the next n-1
# columns of S.  The next n-2 columns are the differences between the
# second column of S and each of the next n-2 columns of S, and so on.
Sdiffs.txt: S.txt
	matrix -col-diffs -in S.txt -out Sdiffs.txt 2>&1 | tee Sdiffs.$(log)

KSdiffs.txt: Sdiffs.txt
	multiply -A K.txt -B Sdiffs.txt -C KSdiffs.txt 2>&1 | tee KSdiffs.$(log)

cand-vertex-basis.txt cand-edge-basis.txt: vertex-basis.txt edge-basis.txt triangle-basis.txt S.txt
	complex -verbose -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt \
	-subcomplex-by-edges S.txt -subcomplex-out-vertex-basis cand-vertex-basis.txt -subcomplex-out-edge-basis cand-edge-basis.txt 2>&1 | tee cand-complex.$(log)

cand-graphviz.dot: cand-vertex-basis.txt cand-edge-basis.txt
	complex -verbose -vertex-basis cand-vertex-basis.txt -edge-basis cand-edge-basis.txt -graphviz cand-graphviz.dot

cand-graphviz.pdf: cand-graphviz.dot
	dot -Tpdf -Kfdp cand-graphviz.dot -o cand-graphviz.pdf

# the following list of columns to prune is created semi-manually at the moment
systolic-coset-gens.txt: S.txt
	matrix -verbose -in S.txt -omit-columns 11,12,14,20,21,22,23,25,27,28,29,30,31,32,33,34,35,37,38,39,41 -out systolic-coset-gens.txt 2>&1 | tee systolic-coset-gens.$(log)

systolic-coset-gens-vertex-basis.txt systolic-coset-gens-edge-basis.txt: vertex-basis.txt edge-basis.txt triangle-basis.txt systolic-coset-gens.txt
	complex -verbose -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt \
	-subcomplex-by-edges systolic-coset-gens.txt -subcomplex-out-vertex-basis systolic-coset-gens-vertex-basis.txt -subcomplex-out-edge-basis systolic-coset-gens-edge-basis.txt 2>&1 | tee systolic-coset-gens-complex.$(log)

systolic-coset-gens-graphviz.dot: systolic-coset-gens-vertex-basis.txt systolic-coset-gens-edge-basis.txt
	complex -verbose -vertex-basis systolic-coset-gens-vertex-basis.txt -edge-basis systolic-coset-gens-edge-basis.txt -graphviz systolic-coset-gens-graphviz.dot

systolic-coset-gens-graphviz.pdf: systolic-coset-gens-graphviz.dot
	dot -Tpdf -Kfdp systolic-coset-gens-graphviz.dot -o systolic-coset-gens-graphviz.pdf


# cochain complex / cosystole targets
#
# we are using the convention here that chain-related items are named
# with subscript index, e.g. d_1, often abbreviated to just d1.  and
# cochain-related items are named with a superscript index, e.g. d^1.
#
# the chain complex is
#
#         d_2          d_1
#   C_2 -------> C_1 -------> C_0
#
# and the cochain complex is
#
#         d^1          d^0
#   C_2 <------- C_1 <------- C_0

d^1.txt: d2.txt
	transpose -in d2.txt -out d^1.txt

d^0.txt: d1.txt
	transpose -in d1.txt -out d^0.txt

d^1smith.txt: d2Tsmith.txt
	cp d2Tsmith.txt d^1smith.txt

d^1colops.txt: d2Tcolops.txt
	cp d2Tcolops.txt d^1colops.txt

d^0smith.txt d^0colops.txt: d^0.txt
	smith -in d^0.txt -smith d^0smith.txt -colops d^0colops.txt 2>&1 | tee d^0smith.$(log)

dim-Z^1.txt: d^1smith.txt
	dim -in d^1smith.txt -kernel > dim-Z^1.txt

dim-B^1.txt: d^0smith.txt
	dim -in d^0smith.txt -image > dim-B^1.txt

dim-H^1.txt: dim-Z^1.txt dim-B^1.txt
	expr `cat dim-Z^1.txt` - `cat dim-B^1.txt` > dim-H^1.txt

# This target computes a basis for Z^1.
Z^1.txt: d^1smith.txt d^1colops.txt dim-Z^1.txt
	automorphism -in d^1colops.txt -dim `dim -in d^1smith.txt -domain` -crop-start `dim -in d^1smith.txt -image` -crop-end `dim -in d^1smith.txt -domain` -out Z^1.txt 2>&1 | tee Z^1.$(log)

# This target checks that the basis for Z^1 we found is actually
# contained in the kernel of d^1.
check-Z^1: d^1Z^1.txt
	rank=$$(dim -in d^1Z^1.txt -image) && if [[ $$rank -eq '0' ]]; then echo "Z^1 is contained in ker d^1"; else echo "Z^1 is NOT contained in ker d^1!"; false; fi

# This target multiplies matrices: d^1 * Z^1, producing d^1Z^1.txt,
# which should be all zeros.
d^1Z^1.txt: d^1.txt Z^1.txt
	multiply -A d^1.txt -B Z^1.txt -C d^1^Z1.txt 2>&1 | tee d^1Z^1.$(log)

# This target computes a basis for the coimage of d^0.
d^0coimage.txt: d^0smith.txt d^0colops.txt
	automorphism -in d^0colops.txt -dim `dim -in d^0smith.txt -domain` -crop-start 0 -crop-end `dim -in d^0smith.txt -image` -out d^0coimage.txt 2>&1 | tee d^0coimage.$(log)

# This target computes a basis for B^1, the image of d^0.
B^1.txt: d^0coimage.txt
	multiply -A d^0.txt -B d^0coimage.txt -C B^1.txt 2>&1 | tee B^1.$(log)

# This target computes a Smith normal form for the matrix B^1 (whose
# columns are a basis for B^1).
B^1smith.txt B^1rowops.txt B^1colops.txt: B^1.txt
	smith -in B^1.txt -smith B^1smith.txt -rowops B^1rowops.txt -colops B^1colops.txt 2>&1 | tee B^1smith.$(log)

# The next three targets are used to compute an aligned basis for B^1
# inside of Z^1.  The resulting basis is stored as columns in
# U^1T.txt.  Note that we call this U^1T because the bras are stored
# as columns, whereas it is more natural to think of them as rows.  We
# later create a file U^1.txt that is the transpose of U^1T.txt.
P^1T.txt: B^1rowops.txt d^1.txt
	automorphism -in B^1rowops.txt -dim `dim -in d^1.txt -domain` -out P^1T.txt 2>&1 | tee P^1T.$(log)

P^1.txt: P^1T.txt
	transpose -in P^1T.txt -out P^1.txt 2>&1 | tee P^1.$(log)

U^1.txt: B^1smith.txt P^1.txt B^1colops.txt Z^1.txt
	align -B1smith B^1smith.txt -P P^1.txt -B1colops B^1colops.txt -Z1 Z^1.txt -out U^1.txt 2>&1 | tee U^1.$(log)

check-U^1: d^1U^1.txt
	rank=$$(dim -in d^1U^1.txt -image) && if [[ $$rank -eq '0' ]]; then echo "U^1 is contained in ker d^1"; else echo "U^1 is NOT contained in ker d^1!"; false; fi

d^1U^1.txt: d^1.txt U^1.txt
	multiply -A d^1.txt -B U^1.txt -C d^1U^1.txt 2>&1 | tee d^1U^1.$(log)

# This target searches for cosystolic vectors in Z^1 \setminus B^1.
cosystoleSample.txt: U^1.txt
	systole -trials 1000 -B B^1.txt -U U^1.txt cosystole.txt 2>&1 | tee cosystoleSample.$(log)

# This target searches for cosystolic vectors in Z^1 \setminus B^1 by the simplicial search algorithm.
cosystole.txt: d1.txt d2.txt Z1.txt
	simplicial-cosystole -d1 d1.txt -d2 d2.txt -z1 Z1.txt 2>&1 | tee cosystole.$(log)

# This target computes the M matrix, which is each cohomology
# generator \bra{\beta}_i evaluated on each homology generator
# \ket{\alpha}_j as the entries of matrix M = (m_{ij}).
M.txt: U^1T.txt U1.txt
	multiply -A U^1T.txt -B U1.txt -C M.txt 2>&1 | tee M.$(log)

U^1T.txt: U^1.txt
	transpose -in U^1.txt -out U^1T.txt 2>&1 | tee U^1T.$(log)

# Smith reduce the M matrix and produce row operations and column
# operations matrices R and C, respectively.  We should have:
#   R M C = I (identity).
Msmith.txt Mrowops.txt Mcolops.txt: M.txt
	smith -in M.txt -smith Msmith.txt -rowops Mrowops.txt -colops Mcolops.txt 2>&1 | tee Msmith.$(log)

R.txt: Mrowops.txt
	automorphism -in Mrowops.txt -transpose -dim `dim -in M.txt -domain` -out R.txt 2>&1 | tee R.$(log)

C.txt: Mcolops.txt
	automorphism -in Mcolops.txt -dim `dim -in M.txt -domain` -out C.txt 2>&1 | tee C.$(log)

# Verify that (R U^1T) (U1 C) = I.
#               F       G
F.txt: R.txt U^1T.txt
	multiply -A R.txt -B U^1T.txt -C F.txt 2>&1 | tee F.$(log)

G.txt: U1.txt C.txt
	multiply -A U1.txt -B C.txt -C G.txt 2>&1 | tee G.$(log)

FG.txt: F.txt G.txt
	multiply -A F.txt -B G.txt -C FG.txt 2>&1 | tee FG.$(log)

check-FG: FG.txt
	rank=$$(dim -in FG.txt -image) && if [[ $$rank -eq `dim -in M.txt -domain` ]]; then echo "FG=I"; else echo "FG does not equal I!"; false; fi

# F should be "reduced" cohomology generators, in that each row of F
# is a cohomology generator that is supported on a single homology
# generator from the G matrix.

simplicial-systole: d1.txt d2.txt
	systole -simplicial -simplicial-at-vertex 0 -simplicial-stop-nonzero -d1 d1.txt -d2 d2.txt


# wip targets
simp-cosys: d1.txt d2.txt Z1.txt B^1.txt U^1.txt
	simplicial-cosystole -verbose -d1 d1.txt -d2 d2.txt -Z1 Z1.txt -B^1 B^1.txt -U^1 U^1.txt 2>&1 | tee simp-cosys.$(log)

simp-cosys-prune: d1.txt d2.txt Z1.txt B^1.txt U^1.txt
	simplicial-cosystole -verbose -prune-by-cohomology-projection -initial-support -d1 d1.txt -d2 d2.txt -Z1 Z1.txt -Z^1 Z^1.txt -B^1 B^1.txt 2>&1 | tee simp-cosys-prune.$(log)

simp-cosys-dim-ucd-seq: d1.txt d2.txt Z1.txt B^1.txt U^1.txt
	simplicial-cosystole -verbose -dim-ucd-seq -d1 d1.txt -d2 d2.txt -Z1 Z1.txt -B^1 B^1.txt -U^1 U^1.txt 2>&1 | tee simp-cosys-dim-ucd-seq.$(log)

simp-cosys-coho-orbits: d1.txt d2.txt Z1.txt B1.txt U^1.txt vertex-basis.txt edge-basis.txt triangle-basis.txt modulus.txt
	simplicial-cosystole -cohomology-orbits -d1 d1.txt -d2 d2.txt -Z1 Z1.txt -B1 B1.txt -U^1 U^1.txt \
		-vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt \
		-modulus `cat modulus.txt` -verbose | tee simp-cosys-coho-orbits.$(log)

coboundary-decoder-threshold: d1.txt d2.txt Z1.txt
	coboundary-decoder -find-threshold -d1 d1.txt -d2 d2.txt -Z_1 Z1.txt -samples-per-weight 10 -verbose -results results-$(log).json 2>&1 | tee coboundary-decoder-threshold.$(log)

circulant:
	random-complex -dimC0 60480 -circulant -regular -k 14 -d1 d1.txt -d2 d2.txt 2>&1 | tee circulant.$(log)

really-clean:
		rm -f B1.txt B1colops.txt B1rowops.txt B1smith.txt C.txt F.txt FG.txt G.txt K.txt KB1.txt KS.txt \
		KSdiffs.txt KSsmith.txt KT.txt KTsmith.txt M.txt Mcolops.txt Mrowops.txt Msmith.txt P1.txt \
		P1T.txt P^1.txt P^1T.txt R.txt S.txt Sdiffs.txt Ssmith.txt U1.txt U^1.txt U^1T.txt Z1.txt Z^1.txt \
		cand-edge-basis.txt cand-graphviz.dot cand-graphviz.pdf cand-vertex-basis.txt cosystole.txt \
		cosystoleExhaustive.txt d1.txt d1S.txt d1U1.txt d1Z1.txt d1colops.txt d1smith.txt d2.txt d2T.txt \
		d2TKT.txt d2Tcolops.txt d2Tsmith.txt d2colops.txt d2smith.txt d^0.txt d^0coimage.txt d^0colops.txt \
		d^0smith.txt d^1.txt d^1U^1.txt d^1Z^1.txt d^1colops.txt d^1smith.txt dim-B1.txt dim-B^1.txt \
		dim-H1.txt dim-H^1.txt dim-Z1.txt dim-Z^1.txt edge-basis.txt results-*.json sample-coboundary-decoder \
		simp-cosys simp-cosys-coho-orbits simp-cosys-dim-ucd-seq simp-cosys-prune systole.txt \
		systolic-coset-gens-graphviz.dot systolic-coset-gens-graphviz.pdf systolic-coset-gens-edge-basis.txt \
		systolic-coset-gens-vertex-basis.txt systolic-coset-gens.txt triangle-basis.txt vertex-basis.txt *.log*


