all: dim-H1.txt systole.txt

log=$(shell logser)
$(info log is $(log))
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
# X, specificaly producing the files:
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
d1.txt d2.txt vertex-basis.txt edge-basis.txt triangle-basis.txt:
	cayley -base-field F4 -d1 d1.txt -d2 d2.txt -vertex-basis vertex-basis.txt -edge-basis edge-basis.txt -triangle-basis triangle-basis.txt 2>&1 | tee cayley.$(log)

# Tool 2.  This tool is more specialized (perhaps convoluted).  It
# produces the basis and boundary files for a Cayley complex, but it
# does so in a completely different manner.  It realizes the
# Cartwright-Steger group, looks for systolic candidate lifts in the
# affine building (infinite complex), then computes the finite
# quotient complex, and projects the systolic candidates down to
# cycles in the finite complex.  So in addition to the basis and
# boundary files, it also produces a systolic candidates file S.txt.
d1.txt d2.txt vertex-basis.txt edge-basis.txt triangle-basis.txt S.txt: modulus.txt
	calg-cayley -modulus `cat modulus.txt` -d1 d1.txt -d2 d2.txt \
		-vertex-basis vertex-basis.txt -edge-basis edge-basis.txt \
		-triangle-basis triangle-basis.txt --systolic-candidates S.txt  2>&1 | tee S.$(log)

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
B1smith.txt B1rowops.txt B1colops: B1.txt
	smith -in B1.txt -smith B1smith.txt -rowops B1rowops.txt -colops B1colops.txt 2>&1 | tee B1smith.$(log)

# The next three targets are used to compute an aligned basis for B_1
# inside of Z_1.  The resulting basis is stored as columns of matrix U
# in U.txt.
PT.txt: B1rowops.txt d1.txt
	automorphism -in B1rowops.txt -dim `dim -in d1.txt -domain` -out PT.txt 2>&1 | tee PT.$(log)

P.txt: PT.txt
	transpose -in PT.txt -out P.txt 2>&1 | tee P.$(log)

U.txt: B1smith.txt P.txt B1colops.txt Z1.txt
	align -B1smith B1smith.txt -P P.txt -B1colops B1colops.txt -Z1 Z1.txt -out U.txt 2>&1 | tee U.$(log)

# The following targets check that the basis for B_1 we found is
# actually contained in the image of d_2.
check-B1: KB1.txt
	rank=$$(dim -in KB1.txt -image) && if [[ $$rank -eq '0' ]]; then echo "B1 is contained in image d2"; else echo "B1 is NOT contained in image d2!"; false; fi

KB1.txt: K.txt B1.txt
	multiply -A K.txt -B B1.txt -C KB1.txt 2>&1 | tee KB1.$(log)

K.txt: KT.txt
	transpose -in KT.txt -out K.txt 2>&1 | tee K.$(log)

KT.txt: d2Tcolops.txt d2Tsmith.txt
	automorphism -in d2Tcolops.txt -dim `dim -in d2Tsmith.txt -domain` -crop-start `dim -in d2Tsmith.txt -image` -out KT.txt 2>&1 | tee KT.$(log)

d2Tsmith.txt d2Tcolops.txt: d2T.txt
	smith -in d2T.txt -smith d2Tsmith.txt -colops d2Tcolops.txt 2>&1 | tee d2Tsmith.$(log)

d2T.txt: d2.txt
	transpose -in d2.txt -out d2T.txt 2>&1 | tee d2T.$(log)

# This target checks that the basis for U we found is actually
# contained in the kernel of d_1.
check-U: d1U.txt KU.txt
	rank=$$(dim -in d1U.txt -image) && if [[ $$rank -eq '0' ]]; then echo "U is contained in ker d1"; else echo "U is NOT contained in ker d1!"; false; fi
	dim -in KU.txt -check-cols-nonzero

d1U.txt: d1.txt U.txt
	multiply -A d1.txt -B U.txt -C d1U.txt 2>&1 | tee d1U.$(log)

KU.txt: K.txt U.txt
	multiply -A K.txt -B U.txt -C KU.txt 2>&1 | tee KU.$(log)

# This target searches for systolic vectors in Z_1 \setminus B_1.
systole.txt: U.txt
	systole -trials 1000 -B B1.txt -U U.txt -min systole.txt 2>&1 | tee systole.$(log)

# Multiplying K by S gives a matrix in which each column corresponds
# to a systolic candidate and each coordinate within the column gives
# a dot product with the corresponding row of K, which is a generator
# for the kernel of d_2^T.  Hence if the column is all zeros, then the
# corresponding candidate is orthogonal to the kernel of d_2^T.  Thus
# by linear algebra, that candidate vector is in the image of d_2.
KS.txt: K.txt S.txt
	multiply -A K.txt -B S.txt -C KS.txt 2>&1 | tee KS.$(log)

rank-KSsmith.txt: KSsmith.txt
	dim -in KSsmith.txt -image > rank-KSsmith.txt

KSsmith.txt: KS.txt
	smith -in KS.txt -smith KSsmith.txt 2>&1 | tee KSsmith.$(log)

# Verify that the systolic candidates are in the kernel of d_1.
d1S.txt: d1.txt S.txt
	multiply -A d1.txt -B S.txt -C d1S.txt 2>&1 | tee d1S.$(log)

check-d1S: d1S.txt
	rank=$$(dim -in d1S.txt -image) && if [[ $$rank -eq '0' ]]; then echo "candidates S are contained in ker d1"; else echo "candidates S are NOT contained in ker d1!"; false; fi


# xxx =============
#
# redo the cosystole targets - do them in the same dir instead of a subdir

init-cosystole:
	mkdir -p cosystole
	transpose -in d1.txt -out cosystole/d2.txt
	cp d2T.txt cosystole/d1.txt
	cp d2Tsmith.txt cosystole/d1smith.txt
	cp d2Tcolops.txt cosystole/d1colops.txt
	cd cosystole && ln -s ../../Makefile .

really-clean:
	rm -f d1smith.txt d1colops.txt d2smith.txt d2colops.txt dim-Z1.txt dim-B1.txt dim-H1.txt Z1.txt d2coimage.txt B1.txt B1smith.txt B1colops.txt Q.txt U.txt KB1.txt K.txt KT.txt d2Tsmith.txt d2Tcolops.txt d2T.txt d1U.txt KU.txt *.log*
