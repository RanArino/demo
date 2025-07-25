## 概要

開発予定のソフトウェアは、大量のテキスト・ドキュメント情報を **情報の粒度（詳細度）** に応じて階層的に構造化し、それを **インタラクティブな2D/3D空間** で視覚化する、ナレッジ探索・情報構造化プラットフォームの開発を目指す。

従来の線形的なテキスト情報（PDF,Markdown,AIチャット履歴など）や、粒度が混在したベクトル検索結果がもたらす情報のオーバーロード問題を解決することが目的。GoogleのNotebookLMのような **高度なRAG（検索拡張生成）技術** と、NotionやObsidianのような **柔軟なドキュメント管理機能** から着想を得て、単なる情報検索を超えた知識の構造化・活用を実現を目指す。

ユーザーは、俯瞰的な視点（ドキュメントの概要やクラスタ）から、段階的に詳細な情報（具体的な文章やチャンク）へとシームレスにドリルダウン可能。さらに、AIチャットの履歴管理を**Gitワークフロー**のような構造で管理し、過去の対話を自由に分岐・編集・統合が可能とする。これにより、ユーザーは思考の変遷や、複数のアイデアを探求が可能となり、**対話そのものが構造化された「記憶」や「思考プロセス」として整理・活用** される。

フロントエンドは人間の直感的な操作と理解に最適化されたインターフェースを提供するが、バックエンドのコア機能は **AIやLLMによる膨大なテキスト情報の自動的な構造化・整理** を主目的として設計する予定。将来的には、AIエージェントが多様なSaaSを連携させる際の **情報ハブ（仲介役）** となり、開発を目指すシステムは「とりあえず情報を渡せば、構造化された知識を出力してくれる」ような役割を担うことを目標とする。

トピック：
- LLM/AI
- SaaS(Software as a Service)
- AaaS(Agent as a Service)
- RAG(Retrieval Augmented Generation)
- Knowledge Graph
- AR(Augmented Reality)

Docs Links:
- [プロジェクトの提案書（日本語）](docs/proposal_ja.md)


## 主な特徴

1.  **階層的な情報構造化:**
    *   取り込んだドキュメント（テキスト、PDF、Webサイト、音声/動画の文字起こし等）をベクトル化し、意味的な関連性に基づいてクラスタリング。
    *   LLMを用いてドキュメントの要約を生成し、「概要」レイヤーと「詳細」レイヤーを構築。
    *   情報の粒度（例: ドキュメント全体 -> 要約クラスタ -> 文章チャンククラスタ -> 個別チャンク）に応じて、情報を複数の階層（レイヤー）に配置。

2.  **インタラクティブな2D/3D可視化:**
    *   階層化された情報構造を、直感的に理解可能な2D/3Dのグラフとして表示。
    *   ユーザーは空間内を自由にナビゲート（回転、ズーム、ノード選択）し、情報の全体像や関連性を視覚的に把握可能。
    *   ノードを選択することで、関連するドキュメント内容のプレビューや詳細情報が表示。

3.  **検索と対話:**
    *   構造化された情報に対して、自然言語での検索（RAG）を実行。
    *   検索結果は、関連する情報がどの階層に存在するかを視覚的に示しながら提示。
    *   AIチャットインターフェースを通じて、ドキュメントの内容に基づいた質疑応答や、対話を通じた知識の深化が可能。対話履歴は分岐・編集可能なフローとして可視化されます。

Docs Links:
- [主要なユーザーインターフェイスの機能概要（英語）](docs/req_frontend.md)
    - ※ 添付画像は、容易にUI構造をイメージするためのものであり、本番環境で実装するものとは異なるので注意
- [システムアーキテクチャ案（英語）](docs/sys_architecture.md)


## 目的
本ソフトウェアは、研究者、学生、ビジネスパーソン、またAI・LLMなど、大量の情報を扱うすべてのユーザーに対して、以下を提供することを目指す：
*   **効率的な知識探索:** 膨大な情報の中から、目的の情報へ直感的かつ迅速なアクセス。
*   **情報理解:** 情報の全体像と詳細、そしてそれらの関連性の構造的な把握。
*   **新たなインサイト発見:** 可視化された情報構造の中から、予期せぬ関連性や新たな知識を発見する機会を創出。
*   **知識基盤:** API や MCP (Message Control Protocol) を通じて、AIが自律的に情報を構造化・検索・理解するためのワークフローコンポーネントとして機能

## Technology Stack & Other Features (Proposed):
### Core Technologies
*   **Frontend:** Next.js (Hybrid Architecture)
    *   **Communication:** Next.js Server Actions (for request-response), gRPC-web (for one-way server streams), WebSockets (for two-way interactive streams).
    *   **UI Components:** Shadcn UI, HeroUI, Tailwind CSS, etc.
    *   **3D Visualization:** react-force-graph-3d, Three.js, etc.
*   **Backend (Polyglot Microservices):**
    *   **Go Services:** For high-performance, low-latency operations (e.g., User Service, API Gateway), CPU-intensive computations, and services requiring minimal memory footprint.
    *   **Python Services:** For AI/ML workloads, data processing (e.g., Knowledge Service, ML Model Serving), and integration with Python-specific libraries.
    *   **Internal Communication:** gRPC (service-to-service).
    *   **External Communication:** gRPC (via Server Actions), gRPC-web (via Envoy proxy), WebSockets (via dedicated service).
*   **Vector Database:** Qdrant
    *   **Purpose:** Vector search, data management.
    *   **Integration:** Go SDK (qdrant-go-client).
*   **LLM & Embedding:**
    *   **Models:** Gemini/Vertex AI (primary), OpenRouter (for diverse model options).
    *   **Purpose:** Text summarization, document segmentation, context understanding.


### 開発者向け機能 (Developer Features)
外部システムやAIエージェントとの連携を可能にするためのバックエンド機能について
*   **gRPC API:**
    *   主要な機能（ドキュメントのアップロード、構造化の開始、検索クエリの実行、ステータス確認など）をプログラムから操作するためのtype-safeで高性能なAPI。
    *   Protocol Bufferによる型安全性とパフォーマンスを提供し、カスタムアプリケーションや他のサービスとの統合を実現。
    *   **Note:** For gRPC-Web proxying, refer to concepts like [Google Cloud Endpoints gRPC Transcoding](https://cloud.google.com/endpoints/docs/grpc/transcoding).
*   **MCP (Message Control Protocol) (検討中):**
    *   AIエージェントが本ソフトウェアの機能をより自律的に利用するための、専用プロトコル（メッセージングベース等を想定）の提供を検討。
    *   AIエージェントが情報を投入し、構造化された結果を受け取り、それに基づいて次のアクションを決定するような、複雑なワークフローを円滑に実行することを想定。

### 拡張機能 (Potential Future Extensions)
*   **ユーザー主体のバージョン:**
    *   主要なバックエンド機能を維持した状態で、研究者、学生、コーポレート、市場関係者など、個々のユーザーに特化したUI設計と機能拡充。
*   **Chrome Extension:**
    *   閲覧中のWebページのテキスト情報を取得し、本ソフトウェアのバックエンドで処理・構造化し、その結果を2D/3Dで即座に表示する機能。
*   **Web AR (拡張現実) 機能:**
    *   スマートフォン等を通じて、QRコードなどをトリガーに、構造化された情報を拡張現実空間に表示する機能。


### クラウド・インフラ
*   **主要クラウド:** AWS / GCP (検討中)
*   **補助クラウド:** Azure (特定機能での利用検討)
*   **インフラ管理:** Terraform (学習目的含む)

### 検討中の外部サービス
*   **認証:** Clerk
*   **サブスクリプション:** Stripe
*   **メール送信:** Resend
