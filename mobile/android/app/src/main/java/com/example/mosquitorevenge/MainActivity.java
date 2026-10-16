package com.example.mosquitorevenge;

import android.app.Activity;
import android.os.Bundle;
import android.view.WindowManager;

import com.example.mosquitorevenge.mobile.EbitenView;
import go.Seq;

public final class MainActivity extends Activity {
    private EbitenView gameView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().setFlags(
                WindowManager.LayoutParams.FLAG_FULLSCREEN,
                WindowManager.LayoutParams.FLAG_FULLSCREEN
        );
        // Harmless with Ebitengine 2.10; needed by older generated mobile bindings.
        Seq.setContext(getApplicationContext());
        gameView = new EbitenView(this);
        setContentView(gameView);
    }

    @Override
    protected void onPause() {
        if (gameView != null) {
            gameView.suspendGame();
        }
        super.onPause();
    }

    @Override
    protected void onResume() {
        super.onResume();
        if (gameView != null) {
            gameView.resumeGame();
        }
    }
}
