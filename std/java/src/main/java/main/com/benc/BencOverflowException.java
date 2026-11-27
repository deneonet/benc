package main.com.benc;

public class BencOverflowException extends BencException {
    public BencOverflowException() {
        super("Varint overflowed a 64-bit integer.");
    }
}