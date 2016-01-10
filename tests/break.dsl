;; simplest

(def i 0)
(for [() true ()]
     (cond (> i 5) (break)
           ())
     (def i (+ i 1))
     )
(assert (== i 6))


;; break and continue
;;(def sum 0)
;;(for [(def i 0) (< i 10) (def i (+ i 1))]
;;     (def sum (+ sum i))
;;     (cond (> i 5) (break)
;;           (continue)))
;;(printf "sum is %d\n" sum)
;;(assert (== sum 15))
;;
;;
