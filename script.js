async function calculateDistance() {
    const address1 = document.getElementById("address1").value;
    const address2 = document.getElementById("address2").value;
    const loadingElement = document.getElementById("loading");
    
    // 検索中メッセージを表示
    loadingElement.style.display = "block";
    document.getElementById("result").innerText = ""; // 結果をクリア

    // 3秒後に処理を実行
    setTimeout(async () => {
        const response = await fetch(`http://shorttransfe.raptech.jp/distance?address1=${encodeURIComponent(address1)}&address2=${encodeURIComponent(address2)}`);
        
        console.log(response); // レスポンスの確認

        if (response.ok) {
            const data = await response.json();
            console.log(data); // データの確認

            if (data.distance_km !== undefined) {
                document.getElementById("result").innerText = `一直線に突っ切る: ${data.distance_km.toFixed(2)} km`;
            } else {
                document.getElementById("result").innerText = `距離情報が取得できませんでした。`;
            }
        } else {
            const errorMessage = await response.text();
            document.getElementById("result").innerText = `エラー: ${errorMessage}`;
        }

        // 処理が終わったら検索中メッセージを隠す
        loadingElement.style.display = "none";
    }, 3000); // 3000ミリ秒（3秒）待つ
}
