# 参考文献・研究資料・その他メモ

## 3D 可視化・階層表現

### Cone Trees: animated 3D visualizations of hierarchical information

**論文URL**: https://dl.acm.org/doi/10.1145/108844.108883

**主要なポイント**:
- この研究にて表現されている階層ツリーは、私自身も再現したいものに近い
- 各ノードに200−300トークン程度のベクトル化されたテキストを配置（キーワードではない）
- ノードを繋ぐ線は、2つの関係性の表現
  - これにて、情報源の保管

**検討事項**:
- 各ノードに対応するベクトルの空間的な配置が、人間が想像する意味的な配置と対応するのかは調査次第
- キーワード類似性 × ベクトル類似性のハイブリットを検討

## 情報可視化・ナビゲーション

### Galaxy of news: an approach to visualizing and understanding expansive news landscapes

**論文URL**: https://dl.acm.org/doi/10.1145/192426.192429

**概要**:
- 独立して作成された大量の情報（ニュース記事）を視覚化するシステム
- 画面をズームすることでより詳細な情報へのアクセスを可能とする
  - 抽象的な情報提示による幅広い情報の探索
  - インタラクションを通じた段階的な詳細情報の理解

**分析・応用案**:
- 自分が目指す理念や方向性とかなり近い
- トピック・タイトルと本文を繋げる中間レイヤー的な存在（学術論文におけるAbstractionやIntroductionに該当）がない
  - **解決策**: AIサマリー生成を導入を検討
- 情報をx, y, zの3D空間で表現されているものの、見えている部分はx-yの平面のみ
  - **目標**: これを"Cone Trees"が見せる角度のようにしたい

### FeedLens: Polymorphic Lenses for Personalizing Exploratory Search over Knowledge Graphs

**論文URL**: https://harmanpk.github.io/Papers/FeedLens_UIST22.pdf

**技術概要**:
- **ポリモーフィックレンズ（FeedLens）**: ユーザーの嗜好モデルを利用して、知識グラフ上の探索的検索結果をランク付けしたりフィルタリング
- パーソナライズされた探索体験を提供するUIST論文で提案された手法

**応用可能性**:
- 情報（ノード）を機械的に制限するアルゴリズムとして使えそう

## 情報理論・哲学的考察

### As We May Think ー 考えてみるに

**文献URL**: https://cruel.org/other/aswemaythink/aswemaythink.pdf

#### 情報圧縮と参照性の課題

**原文抜粋**:
> 人類が活字の発明以来、雑誌、新聞、本、論考、宣伝文句、文通などという形で、記録の総体として十億冊の本に匹敵する記録を生み出したとしたら、そのすべてを集めて圧縮しても、引っ越し用トラックくらいの大きさにして運べる。もちろん圧縮度を高めるだけではダメだ。人は記録を作り保存するだけでなく、それを調べられねばならず、問題のその側面については後で触れる。現代の大図書館ですら、全体が参照されてはいない。少数の人がつまみ食いしているだけなのだ。

**設計への示唆**:
- 後から問題の側面（詳細情報・本質）に戻ることができるように、どのようにトリガーを与えるかがポイント
- **例**: 学術論文のサマリーを読んで、頭の片隅に置いておく → ある時、その論文が論じた問題に高い関連性が予測され、その論文を全て読む
- **トリガーの要素**: その問題が存在するという認知、これに起因する事象間の高い関連性の直感的な予測

**GUI設計原則**:
- ユーザーの質問に対し、まずは広域的な情報を提示し、トリガーを誘発させる（彼らの本質を再確認させる）
- これによって、彼らのクエリの質が高まるとともに、深く探求すべき情報への絞り込みが行われる

#### 情報の位置表現と分類

**原文抜粋**:
> こうした複雑な仕組みが必要なのは私たちが数字を書くのを学んだやり方のせいだ。もしそれを位置として記録し、単なるカード上の点の配置で表せば、自動読み取り装置は比較的簡単になる。実はその点が穴なら、国勢調査のためにホロリスが開発したパンチカード機械がずっと前からあるし、これはいまやあちこちの企業で使われている

**技術的考察**:
- 文字や記号に制限があるが、表現は無限
- **仮説**: 表現された情報を超細分化して、無数だが有限のカテゴリに振り分けた場合、解釈次第では表現の意味的な位置情報を固定できるかもしれない
- **実装案**: 
  - 高次元座標/空間にあらかじめ区画を決めておく
  - 新しいデータが追加された場合には、その区画内で更新・削除を行う
  - これにより、データの圧縮が可能となる

#### 連想的思考とナビゲーション

**原文抜粋**:
> 人がなかなか記録に到達できないのは、索引体系の不自然さから生じている。どんなデータでも保存されると、アルファベット順や数字順に並べられ、情報を見つけるには（見つかればの話だが）それは分類からどんどん下位の分類に下ることになる。それは（複製を使わない限り）たった一つの場所にしかない。どの経路がそれを見つけられるかについてはルールがなければならず、そのルールは面倒なものだ。さらに、一つのアイテムを見つけたら、システムから出てきて、またもや新しい経路に入り直さねばならない。人間の心はそういうふうに機能しない。関連性によって機能する。一つのアイテムをつかんだら、思考の関連性から示唆される次のアイテムに即座にパチッと切り替わる。それは脳細胞が運ぶ何やら複雑な道筋の網目に従ったものだ。

**重要な洞察**:
- **Memexの核心**: 「人間の思考が連想によって機能するという洞察」は、ハイパーテキストやワールドワイドウェブの概念に大きな影響を与えた
- ハイパーテキストは、本質的に情報ノードが相互にリンクされた「空間」を創り出し、ユーザーはその空間を航海する
- これは、ウェブナビゲーションの直接的な前身で、より抽象的には、後にAIによってモデル化される意味空間を航海するという概念の先駆け

**実装アイデア**:
- ユーザーは、情報ノードを行き来することができる仕組み
- **例**: 彼らが10個の情報ノードを航海した場合、5個戻って、そこから異なる航路を進む
- **技術要件**: 5個目までの状態を保存することが重要（例えば、使用された情報ソースやトピック）
- **実現方法**: ユーザーの思考履歴（AIチャットとの対話）をGitのように管理する

## 情報アーキテクチャ設計原則

### Dan Brown's eight design principles lay out the best practices of IA design

**参考URL**:
- https://www.optimalworkshop.com/blog/information-architecture-vs-navigation-creating-a-seamless-user-experience
- https://medium.com/design-bootcamp/dan-browns-eight-useful-principles-of-information-architecture-74caf4df6802

**8つの原則**:

1. **The principle of objects**: Content should be treated as a living, breathing thing. It has lifecycles, behaviors, and attributes.

2. **The principle of choices**: Less is more. Keep the number of choices to a minimum.

3. **The principle of disclosure**: Show a preview of information that will help users understand what kind of information is hidden if they dig deeper.

4. **The principle of examples**: Show examples of content when describing the content of the categories.

5. **The principle of front doors**: Assume that at least 50% of users will use a different entry point than the home page.

6. **The principle of multiple classifications**: Offer users several different classification schemes to browse the site's content.

7. **The principle of focused navigation**: Keep navigation simple and never mix different things.

8. **The principle of growth**: Assume that the content on the website will grow. Make sure the website is scalable.

## 関連サービス・プラットフォーム分析

### InfraNodus

**サービスURL**: https://infranodus.com/

**概要**:
- テキストデータを知識グラフとして視覚化
- 主要なトピックやそれらの関係性、さらには情報間の「ギャップ」を明らかにすることに特化
- グラフを操作しながら思考を深め、AIの支援で新たなアイデアの創出支援

**分析・課題**:
- UIの方向性は似ている
- **問題点**: 情報（ノードとエッジ）を見せすぎで、すべての情報が１つの空間に集約している様子をそのまま表現してしまっている。
- **改善案**: これを制限したい（→階層化）。でなければ、人間の脳はオーバーロードするはず
- **機能拡張の方向性**: 
  - 各ノードには事実を組み込みたい
  - まずはテキスト、のちに画像や動画などへ拡張
- **コンセプトの違い**: "思考を深める"というよりかは、自分のコンセプトは"記憶を辿って事実を引っ張り出す"イメージ

### TheBrain

**サービスURL**: https://www.youtube.com/watch?v=Liy2uJnXg-E

**概要**:
- 情報を関連性に基づいて視覚的なネットワークとして構築
- 人間の思考プロセスを模倣

**分析・評価**:
- **優れた点**: 視覚される情報を制限している点においては、理想に近い
- **UX評価**: ユーザーもあるトピックに関して、ドリルダウンしている感が掴めてそう
- **機能の違い**: ユーザーがノードを追加するという機能は、少し違うかな
- **改善案**: ノードにコメントを追加する方がいいかも

