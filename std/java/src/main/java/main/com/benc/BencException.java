package main.com.benc;

public class BencException extends Exception {
    public BencException(String message) {
        super(message);
    }

    public BencException(String message, Throwable cause) {
        super(message, cause);
    }
}