次をみたすプログラムを書いてください。

なお、言語は Python 、 Ruby 以外の言語であれば何でも良いですが、
Go または Java であるとよいです。

本処理
---

argo-workflows の `WorkflowTemplate` > `spec.templates[*].specs[*][*].arguments.parameters.value` に
渡す値、変数および評価式の linter を作ろうと考えています。

具体的には次の `WorkflowTemplate` リソースのコメントの付与されている部分の linter を作りたいです。
(実際の lint をかける yaml ファイルにはコメントはありません)。
また、 `Workflow` リソースも同様のスキーマを持っており、こちらに対しても同じ linter でチェックできるようにします。

```yaml
apiVersion: argoproj.io/v1alpha1
kind: WorkflowTemplate
metadata:
  name: example
  generateName: example-workflow-
spec:
  entrypoint: example
  templates:
    - name: example
      inputs:
        parameters:
          - name: range-min
          - name: range-max
          - name: option
          - name: expected_duration
            value: ""
      steps:
        - - name: prepare-parameter
            template: prepare-parameter
            arguments:
              parameters:
                - name: min
                  value: "{{ inputs.parameters.range-min }}" # lint
                - name: max
                  value: "{{ inputs.parameters.range-max }}" # lint
                - name: option
                  value: "{{ inputs.parameters.option }}" # lint
        - - name: start-process
            template: start-process
            arguments:
              parameters:
                - name: process-parameter
                  value: "{{ steps.prepare-parameter.outputs.result }}" # lint
        - - name: finish-process
            template: finish-process
            arguments:
              parameters:
                - name: exit-parameter
                  value: "{{=jsonpath(steps['prepare-parameter'].outputs.result, '$.exit')}}" # lint
                - name: exit-parameter
                  value: "{{=jsonpath(steps['prepare-parameter'].outputs.result, '$.exit')}}" # lint
                - name: duration
                  value: "{{=sprig.coalesce(inputs.parameters.expected_duration, steps['start-process'].outputs.parameters.estimated_duration, '1h')}}" # lint
        - - name: notification
            template: notification
            withItems:
              - slack
              - sns
            arguments:
              parameters:
                - name: status
                  value: "{{= steps['start-process'].status == 'Succeeded' && steps['finish-process'].status == 'Succeeded' ? 'Succeeded': 'Failed' }}" # lint
                - name: process-id
                  value: "{{workflow.name}}:{{workflow.uid}}" # lint
                - name: destination
                  value: "{{item}}" # lint
```

`value` に渡せる表現は以下のとおりですが、変数のスキーマについてはドキュメント([Workflow Variables](https://argo-workflows.readthedocs.io/en/latest/variables/#examples))を参照する必要があります。

1. 通常の文字列
2. Workflow Variables(`{{` ではじまり `}}` で終わる部分に記述できる変数)
3. jsonpath 表記(`{{=jsonpath(` で始まる表記。 jsonpath を第二引数に記述できる。 jsonpath の表記が正しいかも検証したい。)
4. sprig 関数(`{{=sprig.xxxx(` で始まる表記。記載された Sprig 関数の存在や、パラメーターの数が正しいかも検証したい。)
5. expr-lang 表記(`{{=` で始まる表記。簡単なフロー関数だけが利用できる。オペレーターが正しいかの検証もしたい。)

上記の 2、3、4、5 の内容が次のいずれかにひっかかる場合にエラーとします。

- 変数の内容を解決できない
- expr-lang の文法として正しくない
- jsonpath のクエリが文法として正しくない
- sprig 関数のシグニチャー(関数名、パラメーター数)が正しくない

プログラム の仕様
---

- プログラムはコマンドラインツールとして開発する
- プログラムの引数にはファイルを渡せるものとします。複数ファイルも可能にできるのであれば、複数ファイルも受け取れるようにしてください。
- lint を実行して、以下の条件の場合はファイル名とリソース名のあとに `ok` とだけ表示します。
    - `WorkflowTemplate` リソースで lint して値の表記に問題がない場合
    - `Workflow` リソースで lint して値の表記に問題がない場合
    - Kubernetes の他のリソースの場合
- lint を実行して、以下の条件の場合はファイル名とリソース名、テンプレート名、ステップ名、パラメーター名、値の内容を表示します。
  - `WorkflowTemplate` リソースで lint して値の表記に問題がある場合
  - `Workflow` リソースで lint して値の表記に問題がある場合
- その他のエラーの場合は、ファイルに対する処理を終了して、次のファイルの処理に移ります。
- 処理をしたファイルのうちに一つ以上エラーがある場合は 0 以外の終了コードにて終了します。

開発の経緯
---

- argo-workflows はワークフローを Kubernetes 上で実現するワークフローエンジンです。
- argo-workflows はワークフローを Kubernetes のカスタムリソースとして記述でき、特定のプログラミング言語を前提としないため学習コストが低くなっています。
- YAML として記述できるので開発の敷居が低いものの、自由度が高くなり、パラメーターの受け渡しに関するデータ形式のチェックの仕組みに弱い部分があります。
- argo lint というツールが公式に提供されており、パラメーターの解決に関するチェック機能は提供されますが、まだまだ必要なチェックが足りていません。
- 具体的に足りていない部分として、今回開発する値の表記に関するチェックや、テンプレートやステップの要求するデータ型のチェックなどがあります。
- 今回開発する lint ツールにより、 argo lint でチェックしきれなかったデータ表記による実行時エラーを回避できることを目指します。
