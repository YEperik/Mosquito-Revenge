package com.example.mosquitorevenge;

import android.app.Activity;
import android.os.Bundle;
import android.view.View;
import android.view.Window;
import android.view.WindowManager;

import com.example.mosquitorevenge.mobile.EbitenView;

public final class MainActivity extends Activity {
    private EbitenView gameView;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        requestWindowFeature(Window.FEATURE_NO_TITLE);
        getWindow().setFlags(
                WindowManager.LayoutParams.FLAG_FULLSCREEN,
                WindowManager.LayoutParams.FLAG_FULLSCREEN
        );
        getWindow().getDecorView().setSystemUiVisibility(
                View.SYSTEM_UI_FLAG_FULLSCREEN
                        | View.SYSTEM_UI_FLAG_HIDE_NAVIGATION
                        | View.SYSTEM_UI_FLAG_IMMERSIVE_STICKY
                        | View.SYSTEM_UI_FLAG_LAYOUT_STABLE
                        | View.SYSTEM_UI_FLAG_LAYOUT_HIDE_NAVIGATION
                        | View.SYSTEM_UI_FLAG_LAYOUT_FULLSCREEN
        );

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
