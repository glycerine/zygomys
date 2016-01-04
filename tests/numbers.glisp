; test different ways of writing an integer
(assert (== 24 0x18))
(assert (== 63 0o77))
(assert (== 13 0b1101))

; test shift operations
(assert (== 4 (sll 1 2)))
(assert (== -1 (sra -4 2)))
(assert (== 2 (srl 4 1)))

; bitwise operations
(assert (== 0b0001 (bit-and 0b0011 0b0101)))
(assert (== 0b0111 (bit-or 0b0011 0b0101)))
(assert (== 0b0110 (bit-xor 0b0011 0b0101)))
(assert (== 0b1100 (bit-and (bit-not 0b0011) 0b1111)))

; arithmetic
(assert (== 5 (+ 3 2)))
(assert (== 2.4 (* 2 1.2)))
(assert (== 2 (mod 5 3)))
(assert (== 1.5 (/ 3 2)))
(assert (== 1.2e3 (* 1.2e2 10)))

(def selection '(1 1.0 0 0.0))

(assert (== '(true true true true) (map number? selection)))
(assert (== '(true false true false) (map int? selection)))
(assert (== '(false true false true) (map float? selection)))
(assert (== '(false false true true) (map zero? selection)))

;; exponentiation
(assert (== 16 (** 2 4)))
(assert (== 1024.0 (** 2.0 10.0)))
