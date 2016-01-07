;; simple for [init test advance]
(def sum 0)
(for [(set i 0) (< i 10) (set i (+ i 1))]
     (set sum (+ sum i)))
(printf "sum is %d\n" sum)
(assert (== sum 45))

