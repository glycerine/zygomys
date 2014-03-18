(def testarr [1 2 3 4 5 6])

(assert (= 3 (aget testarr 2)))
(assert (= 1 (first testarr)))
(assert (= [2 3 4 5 6] (rest testarr)))
(assert (= [1 0 3 4 5 6] (aset! testarr 1 0)))
(assert (= [3 4] (slice testarr 2 4)))
(assert (= [1 2 3] (append [1 2] 3)))
(assert (= [0 1 2 3] (concat [0 1] [2 3])))
