#!/bin/bash

(cd segment && go build && ./segment /tmp/crawl > /tmp/data)
(cd plda && ./lda --num_topics 50 --alpha 0.1 --beta 0.01 --training_data_file /tmp/data --model_file /tmp/model --burn_in_iterations 100 --total_iterations 150 && ./view_model.py /tmp/model > /tmp/readable)
