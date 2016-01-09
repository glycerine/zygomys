;; break and continue
(def sum 0)
(label top
    (for [(def i 0) (< i 10) (def i (+ i 1))]
    (def sum (+ sum i))
    (cond (> i 5) (break top)
         (continue top))))
(printf "sum is %d\n" sum)
(assert (== sum 15))

