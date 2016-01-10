;; break and continue
(def sum 0)
(for [(def i 0) (< i 10) (def i (+ i 1))]
     (def sum (+ sum i))
     ;;(printf "\n wwwwwwhat:  i = %d\n" i)
     (cond (> i 2) (break)
           (continue))
     (def sum (+ sum 1000)) ;; check continue works
     )
(printf "sum is %d\n" sum)
(assert (== sum 6))

