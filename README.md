# myrxgo
仿写rxgo，实现了最基本的功能。
原版rxgo似乎有些地方不尽如人意，复杂而且有很多bug。
so，自己撸了一个。用在项目里，也没出什么乱子（捂脸）。

- map,filter,flatmap等一些基本的operator
- subject，实现了订阅和解除。但是没有实现replay_subject等复杂功能
- 流的clone, merge等操作，但是建议少用，因为这几个功能没有经过项目的验证