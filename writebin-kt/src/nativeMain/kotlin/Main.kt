import platform.posix.putc
import platform.posix.stdout
import kotlin.collections.MutableList

const val EOF : Int = -1;

// kotlin cmdline args do not include the program name
fun main(args: Array<String>) {
    // Join args with comma separation, maybe there's a better way.
    var input = if (args.size > 0) args[0] else ""
    var i = 1
    while (i < args.size) {
        input = input + "," + args[i]
        i++
    }

    // Break apart words.
    //   input ::= numlist?
    //   numlist ::= num (","? num)*
    //   num ::= <something looking like a number>
    var bytes : MutableList<Int> = mutableListOf()
    var (tok0, i0) = next(input, 0)
    if (i0 == EOF) {
        // nothing
    } else if (tok0 == ",") {
        throw Exception("Expected a number at the beginning")
    } else {
        while (true) {
            // invariant: tok0 is a numeric string and i0 is the position after that number
            var n = tok0.toInt()
            if (n > 255) {
                throw Exception("Number too large: " + tok0)
            }
            bytes.add(n);
            var (tok1, i1) = next(input, i0);
            if (i1 == EOF) {
                break
            }
            if (tok1 == ",") {
                var (tok2, i2) = next(input, i1)
                if (i2 == EOF || tok2 == ",") {
                    throw Exception("Expected a number after comma")
                }
                tok1 = tok2
                i1 = i2
            }
            tok0 = tok1
            i0 = i1
        }
    }

    for (n in bytes ) {
        putc(n, stdout)
    }
}

data class NextTok(val item:String, val p:Int);

// Returns (token, nextSourcePosition) or ("", EOF) where token
// is either "," or a valid numeric string.
fun next(input:String, _p:Int): NextTok {
    var p = _p;
    while (p < input.length && input[p] == ' ') {
        p++
    }
    if (p == input.length) {
        return NextTok("", EOF)
    }
    if (input[p] == ',') {
        return NextTok(",", p+1)
    }
    var start = p;
    while (p < input.length && isDigit(input[p])) {
        p++
    }
    if (p == start) {
        throw Exception("Expected number")
    }
    return NextTok(input.substring(start, p), p)
}

fun isDigit(c: Char): Boolean = c >= '0' && c <= '9'
