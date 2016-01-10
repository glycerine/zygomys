;; simplest

(def i 0)
(for [() true ()]
     (cond (> i 5) (break)
           ())
     (def i (+ i 1))
     )
(assert (== i 6))

