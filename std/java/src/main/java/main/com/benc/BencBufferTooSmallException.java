package main.com.benc;

public class BencBufferTooSmallException extends BencException {
    public BencBufferTooSmallException() {
        super("Buffer was too small to complete the operation.");
    }
}